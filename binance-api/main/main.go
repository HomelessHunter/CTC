package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"example.com/wrapper"
	"github.com/gorilla/websocket"
)

func makeWSConnector(dialer *websocket.Dialer, client *http.Client, ctx context.Context) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		conn := wrapper.TickerConnect("btcbusd", dialer, client)

		// go func() {

		// 	for {
		// 		select {
		// 		case <-ctx.Done():
		// 			conn.Close()
		// 			return
		// 		default:
		// 		}
		// 	}

		// }()
		ticker := &wrapper.Ticker{}
		for {
			err := conn.ReadJSON(ticker)
			// _, p, err := conn.ReadMessage()
			if err != nil {
				conn.Close()
				return
			}
			// fmt.Println(len(p))
			fmt.Println(ticker.Data["c"])
		}
	}
}

func makeWSDisconnector(cancel context.CancelFunc) func(http.ResponseWriter, *http.Request) {

	return func(rw http.ResponseWriter, r *http.Request) {
		cancel()
	}
}

func main() {
	dialer := &websocket.Dialer{ReadBufferSize: 256}
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}

	ctx, cancel := context.WithCancel(context.Background())
	mux := http.NewServeMux()
	mux.HandleFunc("/connect", makeWSConnector(dialer, client, ctx))
	mux.HandleFunc("/disconnect", makeWSDisconnector(cancel))
	server := &http.Server{Addr: ":8080", Handler: mux}

	log.Fatal(server.ListenAndServe())
}
