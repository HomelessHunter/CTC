package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	db "github.com/HomelessHunter/CTC/db/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ID int64 = 1

func prepare() (client *mongo.Client, coll *mongo.Collection, err error) {
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	coll = GetUserCollection(client)
	return
}

func closeTest(client *mongo.Client, coll *mongo.Collection) error {
	err := DeleteUserByID(coll, ID, context.TODO())
	if err != nil {
		return err
	}
	client.Disconnect(context.TODO())
	return nil
}

func TestInsertUser(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())

	alertBi, err := db.NewAlert(db.WithPair("btcbusd"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	alertHu, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("huobi"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	user, err := db.NewMongoUser(db.WithUserID(ID), db.WithAlerts(*alertBi, *alertHu))
	if err != nil {
		t.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		t.Errorf("Cannot insert user %s", err)
	}
}

func TestGetUserByID(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())

	user, err := GetUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot get a user %s", err)
	}
	fmt.Println(*user)
}

func TestDeleteUserByID(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())

	err = DeleteUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}

func TestAddAlert(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())

	alert, err := db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(4000.0))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	err = AddAlert(coll, ID, alert, context.TODO())
	if err != nil {
		t.Errorf("Cannot add new alert %s", err)
	}
}

func TestDisctinct(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())

	result, err := GetPairs(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs %s", err)
	}
	fmt.Println(result)
}

func TestRemoveAlert(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer func() {
		err = DeleteUserByID(coll, ID, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		client.Disconnect(context.TODO())
	}()

	alertHu, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("huobi"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	user, err := db.NewMongoUser(db.WithUserID(ID), db.WithAlerts(*alertHu))
	if err != nil {
		t.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		t.Errorf("Cannot insert user %s", err)
	}

	err = RemoveAlert(coll, ID, "btcusdt", context.TODO())
	if err != nil {
		t.Errorf("Cannot remove alert %s", err)
	}

	user, err = GetUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot get a user %s", err)
	}

	if len(user.Alerts) > 0 {
		t.Errorf("lenght of alerts should be 0 but %d instead", len(user.Alerts))
	}

	fmt.Println(user)

}

func TestMarketExist(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer func() {
		err = DeleteUserByID(coll, ID, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		err = DeleteUserByID(coll, 2, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		client.Disconnect(context.TODO())
	}()

	insertUser := func(id int64, market string) {
		alertBi, err := db.NewAlert(db.WithPair("ethusdt"), db.WithTargetPrice(54000.0), db.WithMarket(market))
		if err != nil {
			t.Errorf("Cannot create alert %s", err)
		}
		alertHu, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket(market))
		if err != nil {
			t.Errorf("Cannot create alert %s", err)
		}
		user, err := db.NewMongoUser(db.WithUserID(id), db.WithAlerts(*alertBi, *alertHu))
		if err != nil {
			t.Errorf("Cannot create user %s", err)
		}
		err = InsertNewUser(coll, user, context.TODO())
		if err != nil {
			t.Errorf("Cannot insert user %s", err)
		}
	}

	insertUser(ID, "huobi")
	insertUser(2, "binance")

	exist, err := MarketExist(coll, ID, "binance", context.TODO())
	if err != nil {
		t.Errorf("Cannot check if market exist: %s", err)
	}
	fmt.Println(exist)

}

func prepareForPairsTest(coll *mongo.Collection) error {
	alertBi, err := db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(54000.0), db.WithMarket("binance"), db.WithConnected(true))
	if err != nil {
		return fmt.Errorf("Cannot create alert %s", err)
	}
	alertBin, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"), db.WithConnected(true))
	if err != nil {
		return fmt.Errorf("Cannot create alert %s", err)
	}

	alertHuo, err := db.NewAlert(db.WithPair("ethusdt"), db.WithTargetPrice(54000.0), db.WithMarket("huobi"))
	if err != nil {
		return fmt.Errorf("Cannot create alert %s", err)
	}

	alertBina, err := db.NewAlert(db.WithPair("solusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	if err != nil {
		return fmt.Errorf("Cannot create alert %s", err)
	}

	user, err := db.NewMongoUser(db.WithUserID(ID), db.WithAlerts(*alertBi, *alertBin, *alertHuo, *alertBina), db.WithChatID(1))
	if err != nil {
		return fmt.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		return fmt.Errorf("Cannot insert user %s", err)
	}
	return nil
}

func TestGetPairsByMarket(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}

	defer func() {
		err = DeleteUserByID(coll, ID, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		client.Disconnect(context.TODO())
	}()

	err = prepareForPairsTest(coll)

	pairs, alerts, err := GetPairsByMarket(coll, ID, "binance", true, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs: %s", err)
	}
	fmt.Println(pairs)
	fmt.Println(alerts)

	if cap(pairs) > 2 && cap(alerts) > 2 {
		t.Error("cap greater than 2")
	}
}

func TestUpdateAlerts(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err = closeTest(client, coll)
		if err != nil {
			t.Error(err)
		}
	}()

	err = prepareForPairsTest(coll)
	_, oldAlerts, err := GetPairsByMarket(coll, ID, "binance", false, context.TODO())
	if err != nil {
		t.Error(err)
	}

	newAlerts := make([]db.Alert, len(oldAlerts))
	for i, oldAlert := range oldAlerts {
		oldAlert.Connected = true
		newAlerts[i] = oldAlert
	}

	err = UpdateAlerts(coll, ID, oldAlerts, newAlerts, context.TODO())
	if err != nil {
		t.Error(err)
	}

	alerts, err := GetAlerts(coll, ID, context.TODO())
	if err != nil {
		t.Error(err)
	}
	if !alerts[len(alerts)-1].Connected {
		t.Errorf("connected field should be true but %v instead", alerts[len(alerts)-1].Connected)
	}

	fmt.Println(alerts)
}

func TestDisconnectAlerts(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err = DeleteUserByID(coll, 2, context.TODO())
		if err != nil {
			t.Error(err)
		}
		err = closeTest(client, coll)
		if err != nil {
			t.Error(err)
		}
	}()

	// alertBi, _ := db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(54000.0), db.WithMarket("binance"), db.WithConnected(true))
	// alertBin, _ := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"), db.WithConnected(true))
	// alertHuo, _ := db.NewAlert(db.WithPair("ethusdt"), db.WithTargetPrice(54000.0), db.WithMarket("huobi"))
	// alertBina, _ := db.NewAlert(db.WithPair("solusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	// user, _ := db.NewMongoUser(db.WithUserID(ID), db.WithAlerts(*alertBi, *alertBin, *alertHuo, *alertBina), db.WithChatID(1))
	// err = InsertNewUser(coll, user, context.TODO())
	// if err != nil {
	// 	t.Error(err)
	// }

	alertBi2, _ := db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(4000.0), db.WithMarket("binance"), db.WithConnected(true))
	alertBin2, _ := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"), db.WithConnected(true))
	alertHuo2, _ := db.NewAlert(db.WithPair("ethusdt"), db.WithTargetPrice(4200.0), db.WithMarket("huobi"), db.WithConnected(true))
	alertBina2, _ := db.NewAlert(db.WithPair("solusdt"), db.WithTargetPrice(540.0), db.WithMarket("binance"), db.WithConnected(true))
	user2, _ := db.NewMongoUser(db.WithUserID(2), db.WithAlerts(*alertBi2, *alertBin2, *alertHuo2, *alertBina2), db.WithChatID(2))
	err = InsertNewUser(coll, user2, context.TODO())
	if err != nil {
		t.Error(err)
	}
	sessionAlerts := map[int64][]db.Alert{2: {*alertBi2, *alertBin2, *alertHuo2, *alertBina2}}

	err = ShutdownSequence(coll, sessionAlerts, 8, context.TODO())
	if err != nil {
		t.Error(err)
	}

	// firstUser, err := GetUserByID(coll, 1, context.TODO())
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(firstUser.Alerts)
	secondUser, err := GetUserByID(coll, 2, context.TODO())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(secondUser.Alerts)
}

func TestDeletePairs(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = DeleteUserByID(coll, ID, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		client.Disconnect(context.TODO())
	}()

	err = prepareForPairsTest(coll)
	if err != nil {
		t.Error(err)
	}

	_, alerts, err := GetPairsByMarket(coll, ID, "binance", true, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs: %s", err)
	}
	fmt.Println(alerts)
	err = DeleteAlerts(coll, ID, alerts, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete pairs: %s", err)
	}

	_, alerts, err = GetPairsByMarket(coll, ID, "binance", true, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs: %s", err)
	}
	fmt.Println(alerts)
}

func TestWatchForChanges(t *testing.T) {
	client, coll, err := prepare()
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer func() {
		err = DeleteUserByID(coll, ID, context.TODO())
		if err != nil {
			t.Errorf("Cannot delete user %s", err)
		}
		client.Disconnect(context.TODO())
	}()

	alertBi, err := db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	alertBin, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.3), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}

	alertHuo, err := db.NewAlert(db.WithPair("ethusdt"), db.WithTargetPrice(54000.0), db.WithMarket("huobi"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}

	alertBina, err := db.NewAlert(db.WithPair("solusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}

	user, err := db.NewMongoUser(db.WithUserID(ID), db.WithAlerts(*alertBi, *alertBin, *alertHuo, *alertBina), db.WithChatID(1))
	if err != nil {
		t.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		t.Errorf("Cannot insert user %s", err)
	}

	userPool := make(map[int64]db.MongoUser)

	go func() {
		err = WatchForChanges(coll, userPool, context.TODO())
		if err != nil {
			t.Error(err)
		}
	}()

	<-time.After(2 * time.Second)

	err = RemoveAlert(coll, ID, "ethusdt", context.TODO())
	if err != nil {
		t.Error(err)
	}

	<-time.After(2 * time.Second)

	alert, err := db.NewAlert(db.WithPair("bnbbusd"), db.WithTargetPrice(400.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	err = AddAlert(coll, ID, alert, context.TODO())
	if err != nil {
		t.Errorf("Cannot add new alert %s", err)
	}

	<-time.After(2 * time.Second)

	for k, v := range userPool {
		fmt.Printf("Key: %d, Value: %v", k, &v)
	}

}

func TestCh(t *testing.T) {
	type Alert struct {
		connected bool
	}

	alert := &Alert{}

	ch := make(chan int, 2)
	close := make(chan string, 2)

	go func(alert *Alert) {
		<-ch
		fmt.Println("First func")
		if alert.connected {
			return
		}
		alert.connected = true
		close <- "First"
	}(alert)

	go func(alert *Alert) {
		<-ch
		fmt.Println("Second func")
		if alert.connected {
			return
		}
		alert.connected = true
		close <- "Second"
	}(alert)

	go func() {
		ch <- 1
	}()

	go func() {
		ch <- 1
	}()

	fmt.Println(<-close)
	select {
	case s := <-close:
		fmt.Println(s)
	case <-time.After(2 * time.Second):
		fmt.Println("Second close didn't come")
	}
}
