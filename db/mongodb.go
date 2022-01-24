package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	db "github.com/HomelessHunter/CTC/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUserCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("crypto_bot").Collection("users", options.Collection())
}

func InsertNewUser(coll *mongo.Collection, user *db.MongoUser, ctx context.Context) error {
	_, err := coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByID(coll *mongo.Collection, id int64, ctx context.Context) (*db.MongoUser, error) {
	var user db.MongoUser
	err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteUserByID(coll *mongo.Collection, id int64, ctx context.Context) error {
	_, err := coll.DeleteOne(ctx, bson.D{{"_id", id}})
	if err != nil {
		return err
	}
	return nil
}

func AddAlert(coll *mongo.Collection, id int64, alert *db.Alert, ctx context.Context) error {
	_, err := coll.UpdateByID(ctx, id, bson.D{{"$push", bson.D{{"alerts", alert}}}, {"$set", bson.D{{"timestamp", time.Now()}}}})
	if err != nil {
		return err
	}

	return nil
}

func GetAlertsPairs(coll *mongo.Collection, id int64, ctx context.Context) ([]string, error) {
	result, err := coll.Distinct(ctx, "alerts.pair", bson.D{{"_id", id}})
	if err != nil {
		return nil, err
	}
	pairs := fmt.Sprint(result)
	pairs = pairs[1 : len(pairs)-1]

	return strings.Split(pairs, " "), nil
}
