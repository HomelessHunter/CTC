package wrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

func TickerConnect(symbols string, dialer *websocket.Dialer, client *http.Client) *websocket.Conn {
	if !isSymbolsExist(symbols, client) {
		log.Fatal("Symbol doesn't exist")
	}
	// log.Fatal()
	conn, _, err := dialer.Dial(fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s@miniTicker", strings.ToLower(symbols)), nil)
	if err != nil {
		log.Fatal("Dialer_ERR: ", err)
	}

	// fmt.Println("Response", resp)

	return conn
}

func isSymbolsExist(symbols string, client *http.Client) bool {
	avgPrice := AveragePrice(symbols, client)
	return avgPrice.Msg == ""
}

func AveragePrice(symbols string, client *http.Client) AvgPrice {
	resp, _ := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/avgPrice?symbol=%s", strings.ToUpper(symbols)))
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("AvgPrice", err)
	}
	defer resp.Body.Close()
	avgPrice := &AvgPrice{}
	err = json.Unmarshal(data, avgPrice)
	if err != nil {
		log.Fatal("AvgPrice", err)
	}
	return *avgPrice
}
