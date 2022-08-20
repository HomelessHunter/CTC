package models

import "strings"

type TickerHuobi struct {
	Channel      string       `json:"ch"`
	ResGenTime   int          `json:"ts"`
	StreamDataHu StreamDataHu `json:"tick"`
}

func NewTickerHuobi() *TickerHuobi {
	return &TickerHuobi{}
}

func (ticker *TickerHuobi) GetLastPrice() float64 {
	return ticker.StreamDataHu.LastPrice
}

func (ticker *TickerHuobi) GetSymbol() string {
	if ticker.Channel == "" {
		return ""
	}
	symbol := strings.Split(ticker.Channel, ".")[1]
	return symbol
}

type StreamDataHu struct {
	Id        int     `json:"id"`
	Open      float64 `json:"open"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int     `json:"count"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	LastPrice float64 `json:"lastPrice"`
	LastSize  float64 `json:"lastSize"`
}

type LatestTickerHu struct {
	Channel    string       `json:"ch"`
	Status     string       `json:"status"`
	ResGenTime int64        `json:"ts"`
	LatestData LatestDataHu `json:"tick"`
}

func (latestTicker *LatestTickerHu) GetClosePrice() float64 {
	return latestTicker.LatestData.Close
}

type Ping struct {
	Ping int64 `json:"ping"`
}

func NewPing() *Ping {
	return &Ping{}
}

type LatestDataHu struct {
	Id      int     `json:"id"`
	Low     float64 `json:"low"`
	Open    float64 `json:"open"`
	High    float64 `json:"high"`
	Close   float64 `json:"close"`
	Vol     float64 `json:"vol"`
	Amount  float64 `json:"amount"`
	Version int     `json:"version"`
	Count   int     `json:"count"`
}
