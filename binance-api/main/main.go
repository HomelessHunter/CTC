package main

import (
	"fmt"

	"example.com/wrapper"
)

// func test(func(*websocket.Conn) (int, []byte, error)) {

// }

func main() {

	conn := wrapper.TickerConnect("btcbusd")

	for {
		// ticker := &wrapper.Ticker{}
		// err := conn.ReadJSON(ticker)
		_, p, err := conn.ReadMessage()

		if err != nil {
			conn.Close()
			conn = wrapper.TickerConnect("btcbusd")
		}
		fmt.Println(len(p))
		// fmt.Println(ticker.Data["c"])
	}
}
