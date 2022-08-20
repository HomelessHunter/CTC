package wrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cryptoMarkets "github.com/HomelessHunter/CTC/wrapper/models/cryptoMarkets"
	"github.com/gorilla/websocket"
)

const Huobi string = "huobi"
const Binance string = "binance"

var ErrEmptyPing = errors.New("ping is 0")

func TickerConnect(market string, pairs []string, dialer *websocket.Dialer, client *http.Client) (*websocket.Conn, error) {

	if len(pairs) == 0 {
		return nil, errors.New("pairs shouldn't be empty")
	}

	switch market {
	case Huobi:
		return ConnectHuobi(dialer, pairs)
	case Binance:
		return ConnectBinance(dialer, pairs)
	}

	return nil, fmt.Errorf("no suck market %s", market)
}

func Subscribe(conn *websocket.Conn, pair string, market string) error {
	switch market {
	case Huobi:
		err := SubscribeHu(conn, pair)
		if err != nil {
			return err
		}
	case Binance:
		err := SubscribeBi(conn, pair)
		if err != nil {
			return err
		}
	}
	return nil
}

func Unsubscribe(conn *websocket.Conn, pair string, market string) error {
	switch market {
	case Huobi:
		err := UnsubHu(conn, pair)
		if err != nil {
			return err
		}
	case Binance:
		err := UnsubBi(conn, pair)
		if err != nil {
			return err
		}
	}
	return nil
}

func ConnectHuobi(dialer *websocket.Dialer, pairs []string) (*websocket.Conn, error) {
	conn, _, err := dialer.Dial("wss://api-aws.huobi.pro/ws", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Huobi_Dialer_ERR: %s", err)
		return nil, err
	}
	for _, v := range parsePairsHu(pairs) {
		err = conn.WriteJSON(map[string]string{
			"sub": v,
			"id":  v,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Huobi_WriteJSON: %s", err)
			return nil, err
		}
	}
	return conn, nil
}

func CheckPingHuobi(conn *websocket.Conn, data []byte) error {
	ping := cryptoMarkets.NewPing()
	err := json.Unmarshal(data, ping)
	if err != nil {
		return err
	}
	if ping.Ping > 0 {
		pongMsg := fmt.Sprintf("{\"pong\": %d}", ping.Ping)
		conn.WriteMessage(websocket.TextMessage, []byte(pongMsg))
	} else {
		return ErrEmptyPing
	}
	return nil
}

func SubscribeHu(conn *websocket.Conn, pair string) error {
	err := conn.WriteJSON(map[string]string{
		"sub": parsePairHu(pair),
		"id":  pair,
	})
	if err != nil {
		return err
	}
	return nil
}

func UnsubHu(conn *websocket.Conn, pair string) error {
	err := conn.WriteJSON(map[string]string{
		"unsub": parsePairHu(pair),
		"id":    pair,
	})
	if err != nil {
		return err
	}
	return nil
}

func parsePairsHu(pairs []string) []string {
	for i, v := range pairs {
		pairs[i] = parsePairHu(v)
	}
	return pairs
}

func parsePairHu(pair string) string {
	return fmt.Sprintf("market.%s.ticker", pair)
}

func ConnectBinance(dialer *websocket.Dialer, pairs []string) (*websocket.Conn, error) {
	conn, _, err := dialer.Dial("wss://stream.binance.com:9443/stream", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Binance_Dialer_ERR: %s", err)
		return nil, err
	}

	err = conn.WriteJSON(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": parsePairsBi(pairs),
		"id":     0,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Binance_WriteJSON: %s", err)
		return nil, err
	}
	return conn, nil
}

func SubscribeBi(conn *websocket.Conn, pair string) error {
	err := conn.WriteJSON(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{parsePairBi(pair)},
		"id":     0,
	})
	if err != nil {
		return err
	}
	return nil
}

func UnsubBi(conn *websocket.Conn, pair string) error {
	err := conn.WriteJSON(map[string]interface{}{
		"method": "UNSUBSCRIBE",
		"params": []string{parsePairBi(pair)},
		"id":     0,
	})
	if err != nil {
		return err
	}
	return nil
}

func parsePairsBi(pairs []string) []string {
	for i, v := range pairs {
		pairs[i] = parsePairBi(v)
	}
	return pairs
}

func parsePairBi(pair string) string {
	return fmt.Sprintf("%s@ticker", pair)
}

func getLatestPrice(pair string, client *http.Client) (float64, string, error) {
	latestPriceHu, err := LatestPriceHu(pair, client)
	if err == nil && latestPriceHu.Status != "error" {
		return latestPriceHu.GetClosePrice(), Huobi, nil
	}

	latestPriceBi, err := LatestPriceBi(pair, client)
	if err == nil && latestPriceBi.Msg == "" {
		price, err := latestPriceBi.GetLastPrice()
		if err != nil {
			return 0, "", fmt.Errorf("GetLatestPrice: %s", err)
		}
		return price, Binance, nil
	}

	return 0, "", fmt.Errorf("no data on this pair: %s", pair)
}

func getMarket(pair string, client *http.Client) (string, error) {

	latestPriceHu, err := LatestPriceHu(pair, client)
	if err == nil && latestPriceHu.Status != "error" {
		return Huobi, nil
	}

	latestPriceBi, err := LatestPriceBi(pair, client)
	if err == nil && latestPriceBi.Msg == "" {
		return Binance, nil
	}

	return "", fmt.Errorf("no data on this pair: %s", pair)
}

func getData(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getData: %s", err)
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "getData: %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	return data, nil
}

func LatestPriceBi(pair string, client *http.Client) (*cryptoMarkets.LatestTickerBi, error) {
	data, err := getData(client, fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", strings.ToUpper(pair)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "LatestPriceBi: %s", err)
		return nil, err
	}
	latestPrice := &cryptoMarkets.LatestTickerBi{}
	err = json.Unmarshal(data, latestPrice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LatestPriceBi: %s", err)
		return nil, err
	}
	return latestPrice, nil
}

func LatestPriceHu(pairs string, client *http.Client) (*cryptoMarkets.LatestTickerHu, error) {
	data, err := getData(client, fmt.Sprintf("https://api.huobi.pro/market/detail?symbol=%s", strings.ToLower(pairs)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "LatestPriceHu: %s", err)
		return nil, err
	}
	latestPrice := &cryptoMarkets.LatestTickerHu{}
	err = json.Unmarshal(data, latestPrice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LatestPriceHu: %s", err)
		return nil, err
	}
	return latestPrice, nil
}
