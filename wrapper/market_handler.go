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

func TickerConnect(market string, pairs []string, dialer *websocket.Dialer, client *http.Client) (*websocket.Conn, error) {

	if len(pairs) == 0 {
		return nil, errors.New("pairs shouldn't be empty")
	}

	// if market, err = GetMarket(pairs[len(pairs)-1], client); err != nil {
	// 	fmt.Fprint(os.Stderr, "Symbol doesn't exist")
	// 	return nil, "", err
	// }

	switch market {
	case Huobi:
		return ConnectHuobi(dialer, pairs)
	case Binance:
		return ConnectBinance(dialer, pairs)
	}

	return nil, fmt.Errorf("no suck market %s", market)
}

func ConnectHuobi(dialer *websocket.Dialer, pairs []string) (*websocket.Conn, error) {
	conn, _, err := dialer.Dial("wss://api.huobi.pro/ws", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Huobi_Dialer_ERR: %s", err)
		return nil, err
	}
	for _, v := range parsePairsHu(pairs) {
		err = conn.WriteJSON(map[string]string{
			"sub": v,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Huobi_WriteJSON: %s", err)
			return nil, err
		}
	}
	return conn, nil
}

func ConnectBinance(dialer *websocket.Dialer, pairs []string) (*websocket.Conn, error) {
	conn, _, err := dialer.Dial(fmt.Sprintf("wss://stream.binance.com:9443/stream?streams=%s", parsePairsBi(pairs)), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Binance_Dialer_ERR: %s", err)
		return nil, err
	}
	return conn, nil
}

func parsePairsHu(pairs []string) []string {
	for i, v := range pairs {
		pairs[i] = fmt.Sprintf("market.%s.ticker", v)
	}
	return pairs
}

func parsePairsBi(pairs []string) string {
	if len(pairs) == 1 {
		return pairs[0] + "@miniTicker"
	} else {
		return strings.Join(pairs, "@miniTicker/") + "@miniTicker"
	}
}

func GetMarket(pair string, client *http.Client) (string, error) {

	latestPriceHu, err := LatestPriceHu(pair, client)
	if err == nil && latestPriceHu.Status != "error" {
		return Huobi, nil
	}

	avgPriceBi, err := LatestPriceBi(pair, client)
	if err == nil && avgPriceBi.Msg == "" {
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
		fmt.Fprintf(os.Stderr, "AvgPriceBi: %s", err)
		return nil, err
	}
	avgPrice := &cryptoMarkets.LatestTickerBi{}
	err = json.Unmarshal(data, avgPrice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "AvgPriceBi: %s", err)
		return nil, err
	}
	return avgPrice, nil
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
