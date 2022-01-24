package wrapper

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	cryptoMarkets "github.com/HomelessHunter/CTC/wrapper/models/cryptoMarkets"
	"github.com/gorilla/websocket"
)

func TestConnectHuobi(t *testing.T) {

	dialer := &websocket.Dialer{
		NetDialContext:   (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   256,
		WriteBufferSize:  256,
	}

	conn, _, err := dialer.Dial("wss://api.huobi.pro/ws", nil)
	if err != nil {
		t.Errorf("Huobi_Dialer_ERR_0: %s", err)
	}
	defer conn.Close()

	pairs := []string{"market.btcusdt.ticker", "market.ethusdt.ticker"}

	for _, v := range pairs {
		err = conn.WriteJSON(map[string]string{
			"sub": v,
		})
		if err != nil {
			t.Errorf("Huobi_Dialer_ERR_1: %s", err)
		}
	}

	var buf bytes.Buffer
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("Huobi_Dialer_ERR_2: %s", err)
	}
	buf.Write(data)
	zr, err := gzip.NewReader(&buf)
	if err != nil {
		t.Errorf("Huobi_Dialer_ERR_3: %s", err)
	}
	defer zr.Close()

	ticker := &cryptoMarkets.TickerHuobi{}
	for i := 0; i < 100; i++ {
		zr.Multistream(false)
		_, data, err = conn.ReadMessage()
		if err != nil {
			t.Errorf("Huobi_Dialer_ERR_4: %s", err)
		}
		buf.Write(data)
		data, err = io.ReadAll(zr)
		if err != nil {
			t.Errorf("Huobi_Dialer_ERR_5: %s", err)
		}
		err = json.Unmarshal(data, ticker)
		if err != nil {
			t.Errorf("Huobi_Dialer_ERR_6: %s", err)
		}
		fmt.Println(ticker.GetLastPrice())
		err = zr.Reset(&buf)
		if err != nil {
			t.Errorf("Huobi_Dialer_ERR_7: %s", err)
		}
	}
}

func TestParsePairsHu(t *testing.T) {
	parsedPairs := []string{"market.btcusdt.ticker", "market.ethusdt.ticker"}
	pairs := []string{"btcusdt", "ethusdt"}
	pairs = parsePairsHu(pairs)
	for i, v := range pairs {
		if v != parsedPairs[i] {
			t.Errorf("Slices are not equal")
		}
	}
}
