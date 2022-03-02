package db

import (
	"fmt"
	"testing"
)

func TestAlertSearch(t *testing.T) {
	biAlert := Alert{Market: "binance", Pair: "btcbusd", TargetPrice: 48000.32}
	binAlert := Alert{Market: "binance", Pair: "ethbusd", TargetPrice: 4000.022}
	binaAlert := Alert{Market: "binance", Pair: "solbusd", TargetPrice: 4000.022}
	huAlert := Alert{Market: "binance", Pair: "btcusdt", TargetPrice: 48000.31}
	huoAlert := Alert{Market: "binance", Pair: "ethusdt", TargetPrice: 3000.32}
	huobAlert := Alert{Market: "binance", Pair: "maticusdt", TargetPrice: 3000.32}
	alerts := []Alert{biAlert, binAlert, huAlert, huoAlert, binaAlert, huobAlert}

	testAlert := &Alert{Market: "binance", Pair: "ethbusd"}

	// fmt.Println(testAlert.Pair == binAlert.Pair)

	i, err := testAlert.Find(alerts)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(alerts[i])
}
