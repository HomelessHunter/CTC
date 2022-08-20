package models

import (
	"strconv"
	"strings"
)

type TickerBinance struct {
	Stream string       `json:"stream"`
	Data   StreamDataBi `json:"data"`
}

func NewTickerBi() *TickerBinance {
	return &TickerBinance{}
}

func (ticker *TickerBinance) GetLastPrice() (float64, error) {
	return strconv.ParseFloat(ticker.Data.LastPrice, 64)
}

func (ticker *TickerBinance) GetSymbol() string {
	return strings.ToLower(ticker.Data.Symbol)
}

type StreamDataBi struct {
	Type                string `json:"e"`
	Time                int    `json:"E"`
	Symbol              string `json:"s"`
	PriceChange         string `json:"p"`
	PriceChangePercent  string `json:"P"`
	WeightedAvgPrice    string `json:"w"`
	FirstTrade          string `json:"x"`
	LastPrice           string `json:"c"`
	LastQuantity        string `json:"Q"`
	BestBidPrice        string `json:"b"`
	BestBidQuantity     string `json:"B"`
	BestAskPrice        string `json:"a"`
	BestAskQuantity     string `json:"A"`
	Open                string `json:"o"`
	High                string `json:"h"`
	Low                 string `json:"l"`
	BaseVol             string `json:"v"`
	QuoteVol            string `json:"q"`
	OpenTime            int    `json:"O"`
	CloseTime           int    `json:"C"`
	FirstTradeId        int    `json:"F"`
	LastTradeId         int    `json:"L"`
	TotalNumberOfTrades int    `json:"n"`
}

type LatestTickerBi struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openVolume"`
	CloseTime          int64  `json:"closeVolume"`
	FirstId            int    `json:"firstId"`
	LastId             int    `json:"lastId"`
	Count              int    `json:"count"`
	Code               int    `json:"code"`
	Msg                string `json:"msg"`
}

func (latestTicker *LatestTickerBi) GetLastPrice() (float64, error) {
	return strconv.ParseFloat(latestTicker.LastPrice, 64)
}
