package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/HomelessHunter/CTC/wrapper"
	"github.com/HomelessHunter/CTC/wrapper/models"
	"github.com/gorilla/websocket"
)

var (
	userStreams map[int64]models.UserStreams
	regexps     map[string]*regexp.Regexp
)

func makeWSConnector(dialer *websocket.Dialer, client *http.Client, shutdownSrv chan int) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		wsQuery, err := models.NewWsQuery()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		err = json.Unmarshal(body, wsQuery)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		go alertHandler(dialer, client, wsQuery, shutdownSrv)
	}
}

func alertHandler(dialer *websocket.Dialer, client *http.Client, wsQuery *models.WSQuery, shutdownSrv chan int) error {

	curUserStreams, ok := userStreams[wsQuery.UserId]

	// Check if user have connected previously
	// If false cancel that connection and add a new token pair
	fmt.Println("Before: ", userStreams[wsQuery.UserId])
	if ok {
		// Check if data.Pair already exists
		if pairExist(wsQuery.Pair, curUserStreams.Pairs()) {
			// Send user response
			return errors.New("pair already exists")
		} else {
			curUserStreams.AddPairSignal()
			curUserStreams.Cancel()
			curUserStreams.SetPairs(curUserStreams.AddPairs(strings.ToLower(wsQuery.Pair)))
		}
	} else {
		curUserStreams.SetChatID(wsQuery.ChatId)
		curUserStreams.SetShutdownCh(make(chan int, 1))
		curUserStreams.SetReconnectCh(make(chan int, 1))
		curUserStreams.SetAddPairCh(make(chan int, 1))
		curUserStreams.SetPairs(curUserStreams.AddPairs(strings.ToLower(wsQuery.Pair)))
	}

	fmt.Println("PAIRS: ", curUserStreams.Pairs())

	err := connectToWS(&curUserStreams, dialer, client, wsQuery, shutdownSrv)
	if err != nil {
		return err
	}

	return nil
}

func connectToWS(curUserStreams *models.UserStreams, dialer *websocket.Dialer, client *http.Client, wsQuery *models.WSQuery, shutdownSrv chan int) error {

	if len(curUserStreams.Pairs()) == 0 {
		return wrapper.ErrEmptyPairs
	}

	// Connect to Binance
	conn, err := wrapper.TickerConnect(curUserStreams.Pairs(), dialer, client)
	if err != nil {
		// delete added pair if that's the reason of error
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	curUserStreams.SetCancel(cancel)

	userStreams[wsQuery.UserId] = *curUserStreams
	fmt.Println("After: ", userStreams[wsQuery.UserId])

	// Close websocket connection
	go func(ctx context.Context, shutdownSrv chan int, curUserStreams *models.UserStreams) {
		defer fmt.Println("Websocket CLOSED")
		select {
		case <-shutdownSrv:
			curUserStreams.Shutdown()
			conn.Close()
		case <-ctx.Done():
			fmt.Println("Done")
			select {
			case <-curUserStreams.ReconnectCh():
				curUserStreams.Reconnect()
				conn.Close()
			case <-curUserStreams.ShutdownCh():
				curUserStreams.Shutdown()
				conn.Close()
			case <-curUserStreams.AddPairCh():
				curUserStreams.AddPairSignal()
				conn.Close()
			default:
				err := conn.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Connection closed: %s\n", err)
					return
				}
			}
		}
	}(ctx, shutdownSrv, curUserStreams)

	go checkPrice(conn, dialer, client, wsQuery, shutdownSrv)

	return nil
}

func checkPrice(conn *websocket.Conn, dialer *websocket.Dialer, client *http.Client, wsQuery *models.WSQuery, shutdownSrv chan int) {
	ticker := models.NewTicker()
	defer fmt.Println("checkPrice CLOSED")
	for {
		err := conn.ReadJSON(ticker)

		if err != nil {
			fmt.Println("ReadJson: ", err)
			userStream := userStreams[wsQuery.UserId]

			select {
			case <-userStream.ShutdownCh():
				delete(userStreams, wsQuery.UserId)
				// delete from future db as well
			case <-userStream.ReconnectCh():
				// reconnect
				err := connectToWS(&userStream, dialer, client, wsQuery, shutdownSrv)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case <-userStream.AddPairCh():
				fmt.Println("Adding new pair so do nothing")
			default:
				// Close go-routine to prevent leakage
				userStream.Cancel()
				// reconnect in case of 24h limit or other error
				err := connectToWS(&userStream, dialer, client, wsQuery, shutdownSrv)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			}
			return
		}

		fmt.Println(ticker.GetLastPrice())
	}
}

func makeWSDisconnector() func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		update, err := models.NewCallbackQuery()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		err = json.Unmarshal(body, update)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		disconnectAlert(update)
	}
}

func disconnectAlert(callback *models.CallbackQuery) error {

	elem, ok := userStreams[callback.From.Id]
	if !ok {
		return fmt.Errorf("user with %d ID doesn't have any connections", callback.From.Id)
	}

	// using Callback from TG to get pair for deletion
	// remove pair and reconnect or disconnect completely
	if index := checkDisconnectMsg(callback.Data, elem.Pairs()); index < len(elem.Pairs()) && index >= 0 {
		elem.SetPairs(removePair(elem.Pairs(), index))
		userStreams[callback.From.Id] = elem

		elem.Reconnect()
		elem.Cancel()
	} else {
		elem.Shutdown()
		elem.Cancel()
	}
	return nil
}

func updateFromTG(rw http.ResponseWriter, r *http.Request) {
	// data, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	// update := &wrapper.Update{}
	// err = json.Unmarshal(data, update)
	// if err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// }
	// fmt.Fprintln(os.Stdout, "Update:", update)

	// TEST
	http.Redirect(rw, r, "/disconnect", http.StatusFound)
}

func createUpdate(dialer *websocket.Dialer, client *http.Client, shutdownSrv chan int) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		offset := 0
		id := -1
		for {
			postBody, err := json.Marshal(map[string]int{
				"offset":  -offset,
				"limit":   100,
				"timeout": 0,
			})
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			bodyReq := bytes.NewBuffer(postBody)
			resp, err := http.Post(fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", os.Getenv("TG")), "application/json", bodyReq)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			update := &models.TGUpdate{}
			err = json.Unmarshal(body, update)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			go func(update *models.TGUpdate) {
				if len(update.Result) > 0 {
					result := update.Result[len(update.Result)-1]

					if result.Id == id {
						return
					}

					id = result.Id

					msg := result.Msg

					switch {
					case len(msg.Entities) > 0:
						// handle commands
						fmt.Println(msg.Text)
						us := userStreams[msg.From.Id]
						wsQuery, method := wrapper.CommandRouter(msg.Text, regexps, &result, client, us.Pairs())
						switch method {
						case "alert":
							err := alertHandler(dialer, client, wsQuery, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						default:
							return
						}

					case result.GetCallbackData() != "":

						// handle callbacks
						callback, method := wrapper.CallbackHandler(&result, regexps)

						switch method {
						case "disconnect":
							err := disconnectAlert(callback)

							fmt.Println(callback.Data)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}

							answer, err := models.NewCallbackAnswer(models.WithAnswerID(callback.Id), models.WithAnswerText("Alert(s) disabled"), models.WithAnswerCacheTime(1))
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							wrapper.SendCallbackAnswer(client, answer)
							us := userStreams[callback.From.Id]
							err = wrapper.EditMarkup(client, callback, us.Pairs())
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						default:
							return
						}
					}
				}
			}(update)
		}
	}
}

func main() {
	dialer := &websocket.Dialer{
		NetDialContext:   (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   256,
		WriteBufferSize:  256,
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ReadBufferSize:        256,
			WriteBufferSize:       256,
		},
	}
	shutdownSrv := make(chan int, 1)
	regexps = compileRegexp()
	userStreams = make(map[int64]models.UserStreams)

	mux := http.NewServeMux()
	mux.HandleFunc("/connect", makeWSConnector(dialer, client, shutdownSrv))
	mux.HandleFunc("/disconnect", makeWSDisconnector())
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("TG")), updateFromTG)
	mux.HandleFunc("/update", createUpdate(dialer, client, shutdownSrv))

	server := &http.Server{
		Addr:        ":8080",
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	fmt.Println("Connected")

	if err := server.ListenAndServe(); err != nil {
		shutdownSrv <- 1
		fmt.Println(err)
		client.CloseIdleConnections()
	}

}
