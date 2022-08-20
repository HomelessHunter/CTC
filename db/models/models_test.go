package db

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

var alerts = []Alert{{Market: "binance", Pair: "btcbusd", TargetPrice: 58000, Connected: true, Hex: hex.EncodeToString([]byte("binance" + "btcbusd"))},
	{Market: "huobi", Pair: "ethusdt", TargetPrice: 4500, Connected: true, Hex: hex.EncodeToString([]byte("huobi" + "ethusdt"))},
	{Market: "huobi", Pair: "btcusdt", TargetPrice: 45000, Connected: true, Hex: hex.EncodeToString([]byte("huobi" + "btcusdt"))},
	{Market: "binance", Pair: "solbusd", TargetPrice: 500, Connected: true, Hex: hex.EncodeToString([]byte("binance" + "solbusd"))},
}

func TestAlertSearch(t *testing.T) {

	// alerts := []Alert{{Market: "binance", Pair: "btcbusd", TargetPrice: 58000, Connected: true, Hex: hex.EncodeToString([]byte("binance" + "btcbusd"))},
	// 	{Market: "huobi", Pair: "ethusdt", TargetPrice: 4500, Connected: true, Hex: hex.EncodeToString([]byte("huobi" + "ethusdt"))},
	// 	// {Market: "huobi", Pair: "btcusdt", TargetPrice: 45000, Connected: true, Hex: hex.EncodeToString([]byte("huobi" + "btcusdt"))},
	// 	// {Market: "binance", Pair: "solbusd", TargetPrice: 500, Connected: true, Hex: hex.EncodeToString([]byte("binance" + "solbusd"))},
	// }

	// sort.S
	testAlert := &Alert{Hex: hex.EncodeToString([]byte("huobi" + "ethusdt"))}

	// fmt.Println(testAlert.Pair == binAlert.Pair)
	// sort.SearchStrings()
	SortByHEX(alerts)
	i, err := testAlert.Find(alerts)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(alerts[i])
}

func TestAlertExist(t *testing.T) {
	SortByMarket(alerts)
	if AlertExist(alerts, "binance") && AlertExist(alerts, "huobi") {
		fmt.Println(true)
	} else {
		t.Error("Alert with this market doesn't exist")
	}
}

func TestAlert(t *testing.T) {
	biAlert := Alert{Market: "binance", Pair: "btcbusd", TargetPrice: 48000.32}
	binAlert := Alert{Market: "binance", Pair: "ethbusd", TargetPrice: 4000.022}
	alerts := make([]Alert, 1)
	alerts[0] = biAlert
	fmt.Println(alerts[0])
	alert := &alerts[0]
	alert.SetLastSignal(time.Now())
	alerts = append(alerts, binAlert)
	fmt.Println(alerts[0])
}
