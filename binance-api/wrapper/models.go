package wrapper

type Ticker struct {
	Stream string                 `json:"-"`
	Data   map[string]interface{} `json:"data"`
}

type AvgPrice struct {
	Mins  int    `json:"mins"`
	Price string `json:"price"`
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
}
