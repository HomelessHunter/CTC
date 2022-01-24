package db

import (
	"errors"
	"time"
)

type MongoUser struct {
	UsedID    int64     `bson:"_id"`
	ChatID    int64     `bson:"chat_id"`
	Alerts    []Alert   `bson:"alerts"`
	Timestamp time.Time `bson:"timestamp"`
}

func NewMongoUser(opts ...MongoUserOpts) (*MongoUser, error) {
	user := MongoUser{}
	for _, opt := range opts {
		err := opt(&user)
		if err != nil {
			return nil, err
		}
	}
	user.Timestamp = time.Now()

	return &user, nil
}

type MongoUserOpts func(*MongoUser) error

func WithUserID(userId int64) MongoUserOpts {
	return func(mu *MongoUser) error {
		if userId < 0 {
			return errors.New("userId should be positive")
		}

		mu.UsedID = userId
		return nil
	}
}

func WithChatID(chatId int64) MongoUserOpts {
	return func(mu *MongoUser) error {
		if chatId < 0 {
			return errors.New("chatId should be positive")
		}

		mu.ChatID = chatId
		return nil
	}
}

func WithAlerts(alerts ...Alert) MongoUserOpts {
	return func(mu *MongoUser) error {
		if len(alerts) == 0 {
			return errors.New("alerts shouldn't be empty")
		}

		mu.Alerts = alerts
		return nil
	}
}

type Alert struct {
	Market      string  `bson:"market"`
	Pair        string  `bson:"pair"`
	TargetPrice float32 `bson:"target_price"`
}

func NewAlert(opts ...MongoAlertOpts) (*Alert, error) {
	alert := Alert{}
	for _, opt := range opts {
		err := opt(&alert)
		if err != nil {
			return nil, err
		}
	}

	return &alert, nil
}

type MongoAlertOpts func(*Alert) error

func WithMarket(market string) MongoAlertOpts {
	return func(a *Alert) error {
		if market == "" {
			return errors.New("market shouldn't be empty")
		}

		a.Market = market
		return nil
	}
}

func WithPair(pair string) MongoAlertOpts {
	return func(a *Alert) error {
		if pair == "" {
			return errors.New("pair shouldn't be empty")
		}

		a.Pair = pair
		return nil
	}
}

func WithTargetPrice(targetPrice float32) MongoAlertOpts {
	return func(a *Alert) error {
		if targetPrice < 0 {
			return errors.New("targerPrice should be positive")
		}

		a.TargetPrice = targetPrice
		return nil
	}
}
