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

func TestUserDB(t *testing.T) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		t.Errorf("Cannot connect %s", err)
	}
	defer client.Disconnect(context.TODO())
	coll := GetUserCollection(client)
	alert, err := db.NewAlert(db.WithPair("btcbusd"), db.WithTargetPrice(54000.0), db.WithMarket("binance"))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	user, err := db.NewMongoUser(db.WithUserID(1), db.WithAlerts(*alert))
	if err != nil {
		t.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		t.Errorf("Cannot insert user %s", err)
	}
	user, err = GetUserByID(coll, 1, context.TODO())
	if err != nil {
		t.Errorf("Cannot get a user %s", err)
	}
	fmt.Println(*user)

	<-time.After(3 * time.Second)

	alert, err = db.NewAlert(db.WithPair("ethbusd"), db.WithTargetPrice(4000.0))
	if err != nil {
		t.Errorf("Cannot create alert %s", err)
	}
	err = AddAlert(coll, 1, alert, context.TODO())
	if err != nil {
		t.Errorf("Cannot add new alert %s", err)
	}
	user, err = GetUserByID(coll, 1, context.TODO())
	if err != nil {
		t.Errorf("Cannot get a user %s", err)
	}
	fmt.Println(*user)

	err = DeleteUserByID(coll, 1, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}

func TestDisctinct(t *testing.T) {
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
	user, err := db.NewMongoUser(db.WithUserID(1), db.WithAlerts(*alertBi, *alertHu))
	if err != nil {
		t.Errorf("Cannot create user %s", err)
	}
	err = InsertNewUser(coll, user, context.TODO())
	if err != nil {
		t.Errorf("Cannot insert user %s", err)
	}

	result, err := GetAlertsPairs(coll, 1, context.TODO())
	if err != nil {
		t.Errorf("Cannot get pairs %s", err)
	}
	fmt.Println(result)

	err = DeleteUserByID(coll, 1, context.TODO())
	if err != nil {
		t.Errorf("Cannot delete user %s", err)
	}
}
