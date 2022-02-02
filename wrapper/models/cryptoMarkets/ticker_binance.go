package models

type TickerBinance struct {
	Stream string       `json:"stream"`
	Data   StreamDataBi `json:"data"`
}

func NewTickerBi() *TickerBinance {
	return &TickerBinance{}
}

func (ticker *TickerBinance) GetLastPrice() string {
	return ticker.Data.LastPrice
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
