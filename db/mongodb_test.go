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

func TestInsertUser(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

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
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

	user, err := GetUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot get a user %s", err)
	}
	fmt.Println(*user)
}

func TestDeleteUserByID(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

	err = DeleteUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}

func TestAddAlert(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

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

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

	result, err := GetPairs(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs %s", err)
	}
	fmt.Println(result)
}

func TestRemoveAlert(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

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

	err = DeleteUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}

func TestMarketExist(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)

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

	err = DeleteUserByID(coll, ID, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
	err = DeleteUserByID(coll, 2, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}

func TestGetPairsByMarket(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}

	coll := GetUserCollection(client)
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
	alertBin, err := db.NewAlert(db.WithPair("btcusdt"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
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

	pairs, err := GetPairsByMarket(coll, ID, "huobi", context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs: %s", err)
	}
	fmt.Println(pairs)
}

func TestWatchForChanges(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer func() {
		TestDeleteUserByID(t)
		client.Disconnect(context.TODO())
	}()
	coll := GetUserCollection(client)

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

	chngChan := make(chan db.MongoUser, 2)

	go func() {
		err = WatchForChanges(coll, chngChan, context.TODO())
		if err != nil {
			t.Error(err)
		}
	}()

	<-time.After(2 * time.Second)

	err = RemoveAlert(coll, ID, "ethusdt", context.TODO())
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(<-chngChan)

	<-time.After(2 * time.Second)

	alert, err := db.NewAlert(db.WithPair("bnbbusd"), db.WithTargetPrice(400.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	err = AddAlert(coll, ID, alert, context.TODO())
	if err != nil {
		t.Errorf("Cannot add new alert %s", err)
	}
	f, s := <-chngChan, <-chngChan
	fmt.Println(f, s)
}
