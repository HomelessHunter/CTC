package models

import "errors"

type WSQuery struct {
	UserId int64   `json:"user_id"`
	ChatId int64   `json:"chat_id"`
	Market string  `json:"market"`
	Pair   string  `json:"pair"`
	Price  float64 `json:"price"`
}

func NewWsQuery(opts ...WSQueryOpts) (*WSQuery, error) {
	wsQuery := WSQuery{}

	for _, opt := range opts {
		err := opt(&wsQuery)
		if err != nil {
			return nil, err
		}
	}

	return &wsQuery, nil
}

func (wsQuery *WSQuery) UserID() int64 {
	return wsQuery.UserId
}

type WSQueryOpts func(*WSQuery) error

func WithWSUserId(userId int64) WSQueryOpts {
	return func(w *WSQuery) error {
		if userId < 0 {
			return errors.New("userId should be positive")
		}

		w.UserId = userId
		return nil
	}
}

func WithWSChatId(chatId int64) WSQueryOpts {
	return func(w *WSQuery) error {
		if chatId < 0 {
			return errors.New("chatId should be positive")
		}

		w.ChatId = chatId
		return nil
	}
}

func WithWSMarket(market string) WSQueryOpts {
	return func(w *WSQuery) error {
		if market == "" {
			return errors.New("market shouldn't be empty")
		}

		w.Market = market
		return nil
	}
}

func WithWSPair(pair string) WSQueryOpts {
	return func(w *WSQuery) error {
		if pair == "" {
			return errors.New("pair shouldn't be empty")
		}

		w.Pair = pair
		return nil
	}
}

func WithWSPrice(price float64) WSQueryOpts {
	return func(w *WSQuery) error {
		if price < 0 {
			return errors.New("price should be positive")
		}

		w.Price = price
		return nil
	}
}
