package models

type Ticker struct {
	Stream string                 `json:"-"`
	Data   map[string]interface{} `json:"data"`
}

func NewTicker() *Ticker {
	return &Ticker{}
}

func (ticker *Ticker) GetLastPrice() interface{} {
	return ticker.Data["c"]
}

type AvgPrice struct {
	Mins  int    `json:"mins"`
	Price string `json:"price"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}
