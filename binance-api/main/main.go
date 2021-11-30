package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"example.com/wrapper"
	"github.com/gorilla/websocket"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
)

func makeWSConnector(dialer *websocket.Dialer, client *http.Client) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		ctx, cancel = context.WithCancel(context.Background())
		conn := wrapper.TickerConnect("btcbusd", dialer, client)

		go func() {
			<-ctx.Done()
			fmt.Println("Done")
			conn.Close()
		}()

		ticker := &wrapper.Ticker{}
		for {
			err := conn.ReadJSON(ticker)
			// _, p, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("ReadJson: ", err)
				// conn.Close()
				return
			}
			// fmt.Println(len(p))
			fmt.Println(ticker.Data["c"])
		}
	}
}

func makeWSDisconnector() func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		cancel()
	}
}

func main() {
	dialer := &websocket.Dialer{ReadBufferSize: 256}
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

	mux := http.NewServeMux()
	mux.HandleFunc("/connect", makeWSConnector(dialer, client))
	mux.HandleFunc("/disconnect", makeWSDisconnector())
	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Println("Connected")

	log.Fatal(server.ListenAndServe())
}
