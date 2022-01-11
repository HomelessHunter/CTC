package wrapper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/HomelessHunter/CTC/wrapper/models"
	"github.com/gorilla/websocket"
)

func TickerConnect(pairs []string, dialer *websocket.Dialer, client *http.Client) (*websocket.Conn, error) {
	if !symbolsExist(pairs[len(pairs)-1], client) {
		fmt.Fprint(os.Stderr, "Symbol doesn't exist")
	}

	fmt.Println("ConnectWs: ", checkPairs(pairs))

	conn, _, err := dialer.Dial(fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", checkPairs(pairs)), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Dialer_ERR: %s", err)
		return nil, err
	}

	return conn, nil
}

func checkPairs(pairs []string) string {
	if len(pairs) == 1 {
		return pairs[0] + "@miniTicker"
	} else {
		return strings.Join(pairs, "@miniTicker/") + "@miniTicker"
	}
}

func symbolsExist(pair string, client *http.Client) bool {
	avgPrice := AveragePrice(pair, client)
	return avgPrice.Msg == ""
}

func AveragePrice(pair string, client *http.Client) models.AvgPrice {
	resp, _ := client.Get(fmt.Sprintf("https://api.binance.com/api/v3/avgPrice?symbol=%s", strings.ToUpper(pair)))
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AvgPrice: %s", err)
	}
	defer resp.Body.Close()
	avgPrice := &models.AvgPrice{}
	err = json.Unmarshal(data, avgPrice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AvgPrice: %s", err)
	}
	return *avgPrice
}
