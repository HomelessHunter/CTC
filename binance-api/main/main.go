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

	"example.com/wrapper"
	"github.com/gorilla/websocket"
)

var (
	userSteams map[int64]wrapper.UserStreams
	regexps    map[string]*regexp.Regexp
)

func makeWSConnector(dialer *websocket.Dialer, client *http.Client) func(http.ResponseWriter, *http.Request) {
	userSteams = make(map[int64]wrapper.UserStreams)

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
		if pairExist(data.Pair, userSteams[data.UserId].Pairs) {
			// Send user response
			fmt.Println("Pair already exists")
			return
		}

		// Check if user have connected previously
		// If true cancel that connection and add a new token pair
		fmt.Println("Before: ", userSteams[data.UserId])
		elem, ok := userSteams[data.UserId]
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
			fmt.Fprintln(os.Stderr, err)
			return
		}
		ctx, cancel := context.WithCancel(context.Background())

		elem.Ctx = ctx
		elem.Cancel = cancel

		userSteams[data.UserId] = elem
		fmt.Println("After: ", userSteams[data.UserId])

		go func() {
			<-ctx.Done()
			fmt.Println("Done")
			select {
			case <-elem.ReconnectCh:
				elem.ReconnectCh <- 1
				conn.Close()
			default:
				elem.ShutdownCh <- 1
				conn.Close()
			}
		}()

		go checkPrice(conn, rw, r, &elem, data.UserId)
	}
}

func checkPrice(conn *websocket.Conn, rw http.ResponseWriter, r *http.Request, userStream *wrapper.UserStreams, userId int64) {
	ticker := &wrapper.Ticker{}
	for {
		err := conn.ReadJSON(ticker)

		if err != nil {
			fmt.Println("ReadJson: ", err)
			select {
			case <-userStream.ShutdownCh:
				delete(userSteams, userId)
				// delete from future db as well
			case <-userStream.ReconnectCh:
				// reconnect
				http.Redirect(rw, r, "/connect", http.StatusFound)
			default:
				// do nothing
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
		elem, ok := userSteams[update.Msg.From.Id]
		if !ok {
			fmt.Fprintln(os.Stderr, fmt.Errorf("user with %d ID doesn't have any connections", update.Msg.From.Id))
			return
		}

		disconnectCh := disconnectCh
		// using Callback from TG to get pair for deletion
		go checkDisconnectMsg(update.CallbackQuery.Msg.Text, disconnectCh, elem.Pairs)

		// remove pair and reconnect or disconnect completely
		if index := <-disconnectCh; index < len(elem.Pairs) && index >= 0 {
			elem.Pairs = removePair(elem.Pairs, index)
			userSteams[update.Msg.From.Id] = elem
			userSteams[update.Msg.From.Id].ReconnectCh <- 1
			elem.Cancel()
		} else {
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

func update(rw http.ResponseWriter, r *http.Request) {
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

		if update.Result[len(update.Result)-1].Id == id {
			continue
		}

		id = update.Result[len(update.Result)-1].Id

		if len(update.Result) > 0 {
			msg := update.Result[len(update.Result)-1].Msg
			if s := msg.Entities; len(s) > 0 {
				wrapper.CommandRouter(msg.Text, regexps, update.Result[len(update.Result)-1], rw, r)
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
	mux.HandleFunc("/update", update)
	server := &http.Server{Addr: ":8080", Handler: mux}

	regexps = compileRegexp()

	log.Fatal(server.ListenAndServe())
}
