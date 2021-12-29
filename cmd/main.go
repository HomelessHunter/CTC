package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/HomelessHunter/CTC/wrapper"
	"github.com/gorilla/websocket"
)

var (
	userStreams map[int64]wrapper.UserStreams
	regexps     map[string]*regexp.Regexp
)

func makeWSConnector(dialer *websocket.Dialer, client *http.Client) func(http.ResponseWriter, *http.Request) {
	userStreams = make(map[int64]wrapper.UserStreams)

	return func(rw http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		data := &wrapper.WSQuery{}
		err = json.Unmarshal(body, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// Check if data.Pair already exists
		if pairExist(data.Pair, userStreams[data.UserId].Pairs) {
			// Send user response
			fmt.Println("Pair already exists")
			return
		}

		// Check if user have connected previously
		// If true cancel that connection and add a new token pair
		fmt.Println("Before: ", userStreams[data.UserId])
		elem, ok := userStreams[data.UserId]
		if ok {
			elem.Cancel()
			elem.Pairs = append(elem.Pairs, strings.ToLower(data.Pair))
		} else {
			elem.ChatId = data.ChatId
			elem.ShutdownCh = make(chan int, 1)
			elem.ReconnectCh = make(chan int, 1)
			elem.Pairs = append(elem.Pairs, strings.ToLower(data.Pair))
		}

		// Connect to Binance
		conn, err := wrapper.TickerConnect(elem.Pairs, dialer, client)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot connect", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())

		elem.Cancel = cancel

		userStreams[data.UserId] = elem
		fmt.Println("After: ", userStreams[data.UserId])

		// Close websocket connection
		go func(ctx context.Context) {
			<-ctx.Done()
			fmt.Println("Done")
			select {
			case <-elem.ReconnectCh:
				elem.ReconnectCh <- 1
				conn.Close()
			case <-elem.ShutdownCh:
				elem.ShutdownCh <- 1
				conn.Close()
			default:
				err := conn.Close()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Connection closed: %s\n", err)
				}
			}
		}(ctx)

		go checkPrice(conn, dialer, client, data.UserId)
	}
}

func checkPrice(conn *websocket.Conn, dialer *websocket.Dialer, client *http.Client, userId int64) {
	ticker := &wrapper.Ticker{}
	for {
		err := conn.ReadJSON(ticker)

		if err != nil {
			fmt.Println("ReadJson: ", err)
			userStream := userStreams[userId]

			select {
			case <-userStream.ShutdownCh:
				delete(userStreams, userId)
				// delete from future db as well
			case <-userStream.ReconnectCh:
				// reconnect
				conn, err = wrapper.TickerConnect(userStream.Pairs, dialer, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, "checkPrice: cannot reconnect ", err)
				} else {
					continue
				}
			default:
				// Close go-routine to prevent leakage
				userStream.Cancel()
				// reconnect in case of 24h limit or other error
				conn, err = wrapper.TickerConnect(userStream.Pairs, dialer, client)
				if err != nil {
					fmt.Fprintln(os.Stderr, "checkPrice: cannot reconnect ", err)
				} else {
					continue
				}
			}
			return
		}

		fmt.Println(ticker.GetLastPrice())
	}
}

func makeWSDisconnector(disconnectCh chan int) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		update := &wrapper.Update{}
		err = json.Unmarshal(body, update)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		elem, ok := userStreams[update.Msg.From.Id]
		if !ok {
			fmt.Fprintln(os.Stderr, fmt.Errorf("user with %d ID doesn't have any connections", update.Msg.From.Id))
			return
		}

		disconnectCh := disconnectCh
		// using Callback from TG to get pair for deletion
		go checkDisconnectMsg(update.CallbackQuery.Data, disconnectCh, elem.Pairs)

		// remove pair and reconnect or disconnect completely
		if index := <-disconnectCh; index < len(elem.Pairs) && index >= 0 {
			elem.Pairs = removePair(elem.Pairs, index)
			userStreams[update.Msg.From.Id] = elem
			userStreams[update.Msg.From.Id].ReconnectCh <- 1
			elem.Cancel()
		} else {
			userStreams[update.Msg.From.Id].ShutdownCh <- 1
			elem.Cancel()
		}
	}
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

func createUpdate(client *http.Client) func(http.ResponseWriter, *http.Request) {

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

			update := &wrapper.TGUpdate{}
			err = json.Unmarshal(body, update)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			result := update.Result[len(update.Result)-1]

			if result.Id == id {
				continue
			}

			id = result.Id

			if len(update.Result) > 0 {
				msg := result.Msg

				switch {
				case len(msg.Entities) > 0:
					wrapper.CommandRouter(msg.Text, regexps, result, rw, client, userStreams[msg.From.Id].Pairs)
				case result.CallbackQuery.Data != "":
					// handle callbacks
				}
			}
		}
	}
}

func main() {
	dialer := &websocket.Dialer{ReadBufferSize: 256}
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

	// disconnectCh receives index of pair which should be disconnected from the stream
	// or -1 to disconnect completely
	disconnectCh := make(chan int)

	mux := http.NewServeMux()
	mux.HandleFunc("/connect", makeWSConnector(dialer, client))
	mux.HandleFunc("/disconnect", makeWSDisconnector(disconnectCh))
	mux.HandleFunc(fmt.Sprintf("/%s", os.Getenv("TG")), updateFromTG)
	mux.HandleFunc("/update", createUpdate(client))
	server := &http.Server{Addr: ":8080", Handler: mux}

	regexps = compileRegexp()

	log.Fatal(server.ListenAndServe())
}
