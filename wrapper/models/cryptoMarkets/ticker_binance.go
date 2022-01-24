package models

type TickerBinance struct {
	Stream string       `json:"-"`
	Data   StreamDataBi `json:"data"`
}

func NewTickerBi() *TickerBinance {
	return &TickerBinance{}
}

func (ticker *TickerBinance) GetClosePrice() string {
	return ticker.Data.Close
}

type StreamDataBi struct {
	Type     string `json:"e"`
	Time     int    `json:"E"`
	Symbol   string `json:"s"`
	Close    string `json:"c"`
	Open     string `json:"o"`
	High     string `json:"h"`
	Low      string `json:"l"`
	BaseVol  string `json:"v"`
	QuoteVol string `json:"q"`
}

type LatestTickerBi struct {
	Symbol             string `json:"symbol,omitempty"`
	PriceChange        string `json:"priceChange,omitempty"`
	PriceChangePercent string `json:"priceChangePercent,omitempty"`
	WeightedAvgPrice   string `json:"weightedAvgPrice,omitempty"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice,omitempty"`
	LastQty            string `json:"lastQty,omitempty"`
	BidPrice           string `json:"bidPrice,omitempty"`
	BidQty             string `json:"bidQty,omitempty"`
	AskPrice           string `json:"askPrice,omitempty"`
	AskQty             string `json:"askQty,omitempty"`
	OpenPrice          string `json:"openPrice,omitempty"`
	HighPrice          string `json:"highPrice,omitempty"`
	LowPrice           string `json:"lowPrice,omitempty"`
	Volume             string `json:"volume,omitempty"`
	QuoteVolume        string `json:"quoteVolume,omitempty"`
	OpenTime           int64  `json:"openVolume,omitempty"`
	CloseTime          int64  `json:"closeVolume,omitempty"`
	FirstId            int    `json:"firstId,omitempty"`
	LastId             int    `json:"lastId,omitempty"`
	Count              int    `json:"count,omitempty"`
	Code               int    `json:"code,omitempty"`
	Msg                string `json:"msg,omitempty"`
}
