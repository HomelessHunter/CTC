package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	models "github.com/HomelessHunter/CTC/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetUserCollection(client *mongo.Client) *mongo.Collection {
	return client.Database("crypto_bot").Collection("users", options.Collection())
}

func InsertNewUser(coll *mongo.Collection, user *models.MongoUser, ctx context.Context) error {
	_, err := coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func GetUserByID(coll *mongo.Collection, id int64, ctx context.Context) (*models.MongoUser, error) {
	var user models.MongoUser
	err := coll.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteUserByID(coll *mongo.Collection, id int64, ctx context.Context) error {
	_, err := coll.DeleteOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil
}

func AddAlert(coll *mongo.Collection, id int64, alert *models.Alert, ctx context.Context) error {
	_, err := coll.UpdateByID(ctx, id, bson.D{
		primitive.E{Key: "$push", Value: bson.D{primitive.E{Key: "alerts", Value: alert}}},
		primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "timestamp", Value: time.Now().In(time.UTC)}}}})
	if err != nil {
		return err
	}

	return nil
}

func RemoveAlert(coll *mongo.Collection, id int64, pair string, ctx context.Context) error {
	_, err := coll.UpdateByID(ctx, id, bson.D{primitive.E{Key: "$pull", Value: bson.D{primitive.E{Key: "alerts", Value: bson.D{primitive.E{Key: "pair", Value: pair}}}}},
		primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "timestamp", Value: time.Now().In(time.UTC)}}}})
	if err != nil {
		return err
	}

	return nil
}

// Random order
func GetPairs(coll *mongo.Collection, id int64, ctx context.Context) ([]string, error) {
	result, err := coll.Distinct(ctx, "alerts.pair", bson.D{primitive.E{Key: "_id", Value: id}})
	if err != nil {
		return nil, err
	}

	return splitPairs(result), nil
}

func GetPairsByMarket(coll *mongo.Collection, id int64, market string, ctx context.Context) ([]string, error) {
	var user models.MongoUser
	err := coll.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}},
		options.FindOne().SetProjection(bson.D{primitive.E{Key: "_id", Value: 0},
			primitive.E{Key: "chat_id", Value: 0}, primitive.E{Key: "timestamp", Value: 0}})).Decode(&user)
	if err != nil {
		return nil, err
	}
	pairs := make([]string, len(user.Alerts))

	index := 0
	for _, v := range user.Alerts {
		if v.Market == market {
			pairs[index] = v.Pair
			index++
		}
	}

	pairs = pairs[:index]

	return pairs, nil
}

// func removePair(pairs []string, index int) []string {
// 	return append(pairs[:index], pairs[index+1:]...)
// }

func splitPairs(result []interface{}) []string {
	pairs := fmt.Sprint(result)
	pairs = pairs[1 : len(pairs)-1]
	return strings.Split(pairs, " ")
}

func MarketExist(coll *mongo.Collection, id int64, market string, ctx context.Context) (bool, error) {
	result := coll.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}, primitive.E{Key: "alerts.market", Value: market}})
	if err := result.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func getChangeStream(coll *mongo.Collection, ctx context.Context) (*mongo.ChangeStream, error) {
	pipeline := mongo.Pipeline{bson.D{primitive.E{Key: "$match", Value: bson.D{primitive.E{Key: "operationType", Value: "update"}}}}}
	cs, err := coll.Watch(ctx, pipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func WatchForChanges(coll *mongo.Collection, chngChan chan<- models.MongoUser, ctx context.Context) error {
	cs, err := getChangeStream(coll, ctx)
	if err != nil {
		return err
	}
	defer cs.Close(ctx)

	for cs.Next(ctx) {
		var event bson.M
		user, err := models.NewMongoUser()
		if err != nil {
			return err
		}
		err = cs.Decode(&event)
		if err != nil {
			if err == mongo.ErrClientDisconnected {
				return nil
			}
			return fmt.Errorf("cannot decode event: %s", err)
		}
		output, err := bson.Marshal(event["fullDocument"])
		if err != nil {
			return err
		}

		err = bson.Unmarshal(output, user)
		if err != nil {
			return err
		}
		chngChan <- *user
	}
	return nil
}
