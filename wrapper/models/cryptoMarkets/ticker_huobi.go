package models

type TickerHuobi struct {
	Channel      string       `json:"ch"`
	ResGenTime   int          `json:"ts"`
	StreamDataHu StreamDataHu `json:"tick"`
}

func (ticker *TickerHuobi) GetLastPrice() float32 {
	return ticker.StreamDataHu.LastPrice
}

type StreamDataHu struct {
	Id        int     `json:"id,omitempty"`
	Open      float32 `json:"open"`
	Low       float32 `json:"low"`
	Close     float32 `json:"close"`
	Amount    float32 `json:"amount"`
	Vol       float32 `json:"vol"`
	Count     int     `json:"count"`
	Bid       float32 `json:"bid"`
	BidSize   float32 `json:"bidSize"`
	Ask       float32 `json:"ask"`
	AskSize   float32 `json:"askSize"`
	LastPrice float32 `json:"lastPrice"`
	LastSize  float32 `json:"lastSize"`
}

type LatestTickerHu struct {
	Channel    string       `json:"ch"`
	Status     string       `json:"status"`
	ResGenTime int64        `json:"ts"`
	LatestData LatestDataHu `json:"tick"`
}

func (latestTicker *LatestTickerHu) GetClosePrice() float32 {
	return latestTicker.LatestData.Close
}

type LatestDataHu struct {
	Id      int     `json:"id"`
	Low     float32 `json:"low"`
	Open    float32 `json:"open"`
	High    float32 `json:"high"`
	Close   float32 `json:"close"`
	Vol     float32 `json:"vol"`
	Amount  float32 `json:"amount"`
	Version int     `json:"version"`
	Count   int     `json:"count"`
}
