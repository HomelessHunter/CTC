package wrapper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var dialer = websocket.Dialer{ReadBufferSize: 256}

func TickerConnect(symbols string) *websocket.Conn {
	if !isSymbolsExist(symbols) {
		log.Fatal("Symbol doesn't exist")
	}
	log.Fatal()
	lowerSymbols := strings.ToLower(symbols)
	conn, resp, err := dialer.Dial(fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s@miniTicker", lowerSymbols), nil)
	if err != nil {
		log.Fatal("Dialer_ERR: ", err)
	}

	fmt.Println("Response", resp)

	return conn
}

func isSymbolsExist(symbols string) bool {
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, _ := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/avgPrice?symbol=%s", strings.ToUpper(symbols)))
	s, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(s))
	return err == nil
}
