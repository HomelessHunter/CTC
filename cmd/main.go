package main

import (
	"bytes"
	"compress/gzip"
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
	"syscall"
	"time"

	"github.com/HomelessHunter/CTC/db"
	dbModels "github.com/HomelessHunter/CTC/db/models"
	"github.com/HomelessHunter/CTC/wrapper"
	cryptoMarkets "github.com/HomelessHunter/CTC/wrapper/models/cryptoMarkets"
	other "github.com/HomelessHunter/CTC/wrapper/models/other"
	models "github.com/HomelessHunter/CTC/wrapper/models/telegram"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	userChannels map[int64]other.UserChannels
	regexps      map[string]*regexp.Regexp
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	ctx, cancel := context.WithCancel(context.Background())
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
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		panic(err)
	}

	coll := db.GetUserCollection(mongoClient)

	regexps = compileRegexp()
	userChannels = make(map[int64]other.UserChannels)

	mux := http.NewServeMux()
	// mux.HandleFunc("/connect", makeWSConnector(dialer, client, ctx))
	// mux.HandleFunc("/disconnect", makeWSDisconnector())
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("TG")), updateFromTG)
	mux.HandleFunc("/update", createUpdate(dialer, client, ctx, coll))

	server := &http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		// WriteTimeout: 10 * time.Second,
		IdleTimeout: 30 * time.Second,
	}

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, os.Interrupt)
		<-sigs
		db.DeleteUserByID(coll, 1, ctx)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
	}()

	fmt.Println("Connected")

	// TEST
	user, err := dbModels.NewMongoUser(
		dbModels.WithUserID(1),
		dbModels.WithChatID(1))
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot create user")
	}
	db.InsertNewUser(coll, user, ctx)

	if err := server.ListenAndServe(); err != nil {
		mongoClient.Disconnect(ctx)
		cancel()
		client.CloseIdleConnections()
		fmt.Println(err)
	}

}

// func makeWSConnector(dialer *websocket.Dialer, client *http.Client, shutdownSrv context.Context) func(http.ResponseWriter, *http.Request) {
//
// 	return func(rw http.ResponseWriter, r *http.Request) {
// 		body, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 		}
//
// 		wsQuery, err := other.NewWsQuery()
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			return
// 		}
//
// 		err = json.Unmarshal(body, wsQuery)
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			return
// 		}
// 		go alertHandler(dialer, client, wsQuery, shutdownSrv)
// 	}
// }

// WHAT IF USER WANT TO RECEIVE INFO FROM DIFF MARKETS
// THAT WON'T WORK UNLESS YOU CHANGE IT
func alertHandler(dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context, coll *mongo.Collection) error {

	curUserChannel, ok := userChannels[wsQuery.UserId]
	var alert *dbModels.Alert
	var err error
	// Check if user have connected previously
	// If false cancel that connection and add a new token pair
	fmt.Println("Before: ", userChannels[wsQuery.UserId])
	if ok {
		// Check if data.Pair already exists
		if pairExist(wsQuery.Pair, curUserChannel.Pairs()) {
			// Send user response
			return errors.New("pair already exists")
		}

		// DO I NEED AN ARRAY OF CHANNELS OR MAP
		// AT LEAST 2 FOR DIFF CONNECTIONS
		// if MarketExist(wsQuery.Market) {
		// 	curUserChannel.AddPairSignal(wsQuery.Market)
		// 	curUserChannel.Cancel(wsQuery.Market)
		// }

		alert, err = dbModels.NewAlert(dbModels.WithPair(wsQuery.Pair), dbModels.WithTargetPrice(wsQuery.Price))
		if err != nil {
			return fmt.Errorf("cannot create new alert: %s", err)
		}
		db.AddAlert(coll, wsQuery.UserId, alert, shutdownSrv)
		curUserChannel.SetPairs(curUserChannel.AddPairs(strings.ToLower(wsQuery.Pair)))

	} else {
		alert, err = dbModels.NewAlert(dbModels.WithPair(wsQuery.Pair), dbModels.WithTargetPrice(float32(wsQuery.Price)))
		if err != nil {
			return fmt.Errorf("cannot create new alert: %s", err)
		}
		db.AddAlert(coll, wsQuery.UserId, alert, shutdownSrv)
		curUserChannel.SetChatID(wsQuery.ChatId)
		curUserChannel.SetShutdownCh(wsQuery.Market, make(chan int, 1))
		curUserChannel.SetReconnectCh(wsQuery.Market, make(chan int, 1))
		curUserChannel.SetAddPairCh(wsQuery.Market, make(chan int, 1))
		curUserChannel.SetPairs(curUserChannel.AddPairs(strings.ToLower(wsQuery.Pair)))
	}

	fmt.Println("PAIRS: ", curUserChannel.Pairs())

	err = connectToWS(&curUserChannel, dialer, client, wsQuery, shutdownSrv)
	if err != nil {
		return err
	}

	return nil
}

func connectToWS(curUserStreams *other.UserChannels, dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context) error {

	// Connect to Market
	conn, err := wrapper.TickerConnect(wsQuery.Market, curUserStreams.Pairs(), dialer, client)
	if err != nil {
		// delete added pair if that's the reason of error
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	curUserStreams.SetCancel(wsQuery.Market, cancel)

	userChannels[wsQuery.UserId] = *curUserStreams
	fmt.Println("After: ", userChannels[wsQuery.UserId])

	// Close websocket connection
	go func(ctx context.Context, shutdownSrv context.Context, curUserStreams *other.UserChannels) {
		defer fmt.Println("Websocket CLOSED")
		select {
		case <-shutdownSrv.Done():
			curUserStreams.Shutdown(wsQuery.Market)
			conn.Close()
		case <-ctx.Done():
			fmt.Println("Done")
			select {
			case <-curUserStreams.ReconnectCh(wsQuery.Market):
				curUserStreams.Reconnect(wsQuery.Market)
				conn.Close()
			case <-curUserStreams.ShutdownCh(wsQuery.Market):
				curUserStreams.Shutdown(wsQuery.Market)
				conn.Close()
			case <-curUserStreams.AddPairCh(wsQuery.Market):
				curUserStreams.AddPairSignal(wsQuery.Market)
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

func checkPrice(conn *websocket.Conn, dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context) {
	defer fmt.Println("CheckPrice is DONE")
	switch wsQuery.Market {
	case wrapper.Huobi:
		checkPriceHu(conn, dialer, client, wsQuery, shutdownSrv)
	case wrapper.Binance:
		checkPriceBi(conn, dialer, client, wsQuery, shutdownSrv)
	}
}

func checkPriceBi(conn *websocket.Conn, dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context) {
	ticker := cryptoMarkets.NewTickerBi()
	defer fmt.Println("checkPriceBi is DONE")
	for {
		err := conn.ReadJSON(ticker)

		if err != nil {
			fmt.Println("ReadJson: ", err)
			handleConnReadErr(dialer, client, wsQuery, shutdownSrv)
			return
		}

		fmt.Println(ticker.GetClosePrice())
	}
}

func checkPriceHu(conn *websocket.Conn, dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context) {
	defer fmt.Println("CheckPriceHu is DONE")
	var buf bytes.Buffer
	_, data, err := conn.ReadMessage()
	if err != nil {
		handleConnReadErr(dialer, client, wsQuery, shutdownSrv)
		return
	}
	buf.Write(data)
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		handleCheckPriceErr(wsQuery)
		return
	}
	defer zr.Close()

	ticker := &cryptoMarkets.TickerHuobi{}
	for {
		zr.Multistream(false)
		_, data, err = conn.ReadMessage()
		if err != nil {
			handleConnReadErr(dialer, client, wsQuery, shutdownSrv)
			return
		}
		buf.Write(data)
		data, err = io.ReadAll(zr)
		if err != nil {
			handleCheckPriceErr(wsQuery)
			return
		}
		err = json.Unmarshal(data, ticker)
		if err != nil {
			handleCheckPriceErr(wsQuery)
			return
		}
		fmt.Println(ticker.GetLastPrice())
		err = zr.Reset(&buf)
		if err != nil {
			handleCheckPriceErr(wsQuery)
			return
		}
	}
}

func handleCheckPriceErr(wsQuery *other.WSQuery) {
	userChannel := userChannels[wsQuery.UserId]
	userChannel.Cancel(wsQuery.Market)
}

func handleConnReadErr(dialer *websocket.Dialer, client *http.Client, wsQuery *other.WSQuery, shutdownSrv context.Context) {
	userChannel := userChannels[wsQuery.UserId]
	select {
	case <-userChannel.ShutdownCh(wsQuery.Market):
		delete(userChannels, wsQuery.UserId)

	case <-userChannel.ReconnectCh(wsQuery.Market):
		// reconnect
		err := connectToWS(&userChannel, dialer, client, wsQuery, shutdownSrv)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	case <-userChannel.AddPairCh(wsQuery.Market):
		fmt.Println("Adding new pair so do nothing")
	default:
		// Close go-routine to prevent leakage
		userChannel.Cancel(wsQuery.Market)
		// reconnect in case of 24h limit or other error
		err := connectToWS(&userChannel, dialer, client, wsQuery, shutdownSrv)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

// func makeWSDisconnector() func(http.ResponseWriter, *http.Request) {
//
// 	return func(rw http.ResponseWriter, r *http.Request) {
// 		body, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 		}
//
// 		update, err := models.NewCallbackQuery()
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 			return
// 		}
//
// 		err = json.Unmarshal(body, update)
// 		if err != nil {
// 			fmt.Fprintln(os.Stderr, err)
// 		}
//
// 		disconnectAlert(update)
// 	}
// }

func disconnectAlert(coll *mongo.Collection, ctx context.Context, callback *models.CallbackQuery) error {

	pairs, err := db.GetAlertsPairs(coll, callback.From.Id, ctx)
	if err != nil || len(pairs) == 0 {
		return fmt.Errorf("user with %d ID doesn't have any connections", callback.From.Id)
	}
	elem, ok := userChannels[callback.From.Id]
	if !ok {
		return fmt.Errorf("user with %d ID doesn't have any connections", callback.From.Id)
	}
	// split callback data to get pair and market
	pair, market := wrapper.SplitCallbackData(callback.Data)
	// using Callback from TG to get pair for deletion
	// remove pair and reconnect or disconnect completely
	index, err := checkDisconnectMsg(pair, pairs)
	if err != nil {
		return errors.New("cannot check msg")
	}
	if index < len(pairs) && index >= 0 {
		elem.SetPairs(removePair(elem.Pairs(), index))
		userChannels[callback.From.Id] = elem

		elem.Reconnect(market)
		elem.Cancel(market)
	} else {
		elem.Shutdown(market)
		elem.Cancel(market)
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

func createUpdate(dialer *websocket.Dialer, client *http.Client, shutdownSrv context.Context, coll *mongo.Collection) func(http.ResponseWriter, *http.Request) {

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
						pairs, err := db.GetAlertsPairs(coll, result.FromUser().Id, shutdownSrv)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}
						// us := userChannels[msg.From.Id]
						wsQuery, method, err := wrapper.CommandRouter(msg.Text, regexps, &result, client, pairs)
						if err != nil {
							fmt.Fprintln(os.Stderr, fmt.Errorf("cannot parse command: %s", err))
						}

						switch method {
						case "alert":
							err := alertHandler(dialer, client, wsQuery, shutdownSrv, coll)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
						default:
							return
						}

					case result.GetCallbackData() != "":

						// handle callbacks
						callback, method, err := wrapper.CallbackHandler(client, &result, regexps)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
							return
						}

						switch method {
						case "disconnect":
							err := disconnectAlert(coll, shutdownSrv, callback)

							// fmt.Println(callback.Data)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}

							answer, err := models.NewCallbackAnswer(
								models.WithAnswerID(callback.Id),
								models.WithAnswerText("Alert(s) disabled"),
								models.WithAnswerCacheTime(1),
							)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							wrapper.SendCallbackAnswer(client, answer)
							pairs, err := db.GetAlertsPairs(coll, result.FromUser().Id, shutdownSrv)
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							// us := userChannels[callback.From.Id]
							err = wrapper.EditMarkup(client, callback, pairs)
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
