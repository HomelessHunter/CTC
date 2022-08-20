package db

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"
)

type MongoUser struct {
	UsedID    int64     `bson:"_id"`
	ChatID    int64     `bson:"chat_id"`
	Alerts    []Alert   `bson:"alerts"`
	Timestamp time.Time `bson:"timestamp"`
}

func (user *MongoUser) String() string {
	return fmt.Sprintf("UserID: %d\nChatID: %d\nAlerts: %v\nTimestamp: %v\n", user.UsedID, user.ChatID, user.Alerts, user.Timestamp)
}

func NewMongoUser(opts ...MongoUserOpts) (*MongoUser, error) {
	user := MongoUser{}
	for _, opt := range opts {
		err := opt(&user)
		if err != nil {
			return nil, err
		}
	}
	user.Timestamp = time.Now().In(time.UTC)

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
	Market      string    `bson:"market"`
	Pair        string    `bson:"pair"`
	TargetPrice float64   `bson:"target_price"`
	Connected   bool      `bson:"connected"`
	LastSignal  time.Time `bson:"last_signal,omitempty"`
	Hex         string    `bson:"hex"`
}

func (alert *Alert) String() string {
	return fmt.Sprintf("Market: %s, Pair: %s, TargetPrice: %f", alert.Market, alert.Pair, alert.TargetPrice)
}

func NewAlert(opts ...MongoAlertOpts) (*Alert, error) {
	alert := Alert{}
	for _, opt := range opts {
		err := opt(&alert)
		if err != nil {
			return nil, err
		}
	}

	alert.Hex = hex.EncodeToString([]byte(alert.Market + alert.Pair))

	return &alert, nil
}

func (alert *Alert) SetLastSignal(lastSignal time.Time) {
	alert.LastSignal = lastSignal
}

func SortByHEX(alerts []Alert) {
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Hex < alerts[j].Hex
	})
}

func (alert *Alert) SortNFind(alerts []Alert) (int, error) {
	SortByHEX(alerts)
	fmt.Println("SORT", alerts)
	i := sort.Search(len(alerts), func(i int) bool {
		return alerts[i].Hex >= alert.Hex
	})
	if i < len(alerts) && alerts[i].Hex == alert.Hex {
		return i, nil
	}

	return 0, fmt.Errorf("no alert with index: %d", i)
}

func (alert *Alert) Find(alerts []Alert) (int, error) {
	i := sort.Search(len(alerts), func(i int) bool {
		return alerts[i].Hex >= alert.Hex
	})
	if i < len(alerts) && alerts[i].Hex == alert.Hex {
		return i, nil
	}

	return 0, fmt.Errorf("no alert with index: %d", i)
}

func SortByMarket(alerts []Alert) {
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].Market < alerts[j].Market
	})
}

func AlertExist(alerts []Alert, market string) bool {
	i := sort.Search(len(alerts), func(i int) bool {
		return alerts[i].Market >= market
	})
	if i < len(alerts) && alerts[i].Market == market {
		return true
	}
	return false
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

func WithTargetPrice(targetPrice float64) MongoAlertOpts {
	return func(a *Alert) error {
		if targetPrice < 0 {
			return errors.New("targerPrice should be positive")
		}

		a.TargetPrice = targetPrice
		return nil
	}
}

func WithConnected(connected bool) MongoAlertOpts {
	return func(a *Alert) error {
		a.Connected = connected
		return nil
	}
}

func WithLastSignal(lastSignal time.Time) MongoAlertOpts {
	return func(a *Alert) error {
		if lastSignal.IsZero() {
			return errors.New("lastSignal shouldn't be 0")
		}
		a.LastSignal = lastSignal
		return nil
	}
}

func WithHex(hex string) MongoAlertOpts {
	return func(a *Alert) error {
		if hex == "" {
			return errors.New("hex shouldn't be empty")
		}
		a.Hex = hex
		return nil
	}
}
