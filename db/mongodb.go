package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	models "github.com/HomelessHunter/CTC/db/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var NoPairsErr error = errors.New("No pairs associated with this user")

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

func GetUsersWithPairs(coll *mongo.Collection, connected bool, ctx context.Context) ([]models.MongoUser, error) {
	var users []models.MongoUser
	cursor, err := coll.Find(ctx, bson.D{
		primitive.E{
			Key: "alerts.0", Value: bson.D{primitive.E{Key: "$exists", Value: true}},
		},
		primitive.E{
			Key: "alerts", Value: bson.D{primitive.E{Key: "$elemMatch", Value: bson.D{primitive.E{Key: "connected", Value: connected}}}},
		}},
	)
	if err != nil {
		return nil, fmt.Errorf("GetUsersWithPairs: %s", err)
	}
	err = cursor.All(ctx, &users)
	if err != nil {
		return nil, fmt.Errorf("GetUsersWithPairs: %s", err)
	}
	return users, nil
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

func DeleteAlerts(coll *mongo.Collection, id int64, alerts []models.Alert, ctx context.Context) error {
	_, err := coll.UpdateByID(ctx, id, bson.D{primitive.E{Key: "$pullAll", Value: bson.D{primitive.E{Key: "alerts", Value: alerts}}},
		primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "timestamp", Value: time.Now().In(time.UTC)}}}})
	if err != nil {
		return err
	}
	return nil
}

func ShutdownSequence(coll *mongo.Collection, sessionAlerts map[int64]*[]models.Alert, count int, ctx context.Context) error {
	writeModels := make([]mongo.WriteModel, 0, count)
	for k, v := range sessionAlerts {
		writeModels = append(writeModels, setAlertsConnected(k, *v, false)...)
	}

	_, err := coll.BulkWrite(ctx, writeModels)
	if err != nil {
		return fmt.Errorf("ShutdownSequence: %s", err)
	}
	return nil
}

func UpdateAlertsSqc(coll *mongo.Collection, id int64, alerts []models.Alert, connected bool, ctx context.Context) error {
	_, err := coll.BulkWrite(ctx, setAlertsConnected(id, alerts, connected))
	if err != nil {
		return fmt.Errorf("UpdateAlertsSqc: %s", err)
	}
	return nil
}

func setAlertsConnected(id int64, alerts []models.Alert, connected bool) []mongo.WriteModel {
	updates := make([]mongo.WriteModel, len(alerts))
	for i, v := range alerts {
		updates[i] = mongo.NewUpdateOneModel().SetFilter(
			bson.D{primitive.E{Key: "_id", Value: id}, primitive.E{Key: "alerts.pair", Value: v.Pair}},
		).SetUpdate(bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "alerts.$.connected", Value: connected}}}})
	}
	return updates
}

// Random order
func GetPairs(coll *mongo.Collection, id int64, ctx context.Context) ([]string, error) {
	result, err := coll.Distinct(ctx, "alerts.pair", bson.D{primitive.E{Key: "_id", Value: id}})
	if err != nil {
		return nil, err
	}
	return splitPairs(result), nil
}

func GetAlerts(coll *mongo.Collection, id int64, ctx context.Context) ([]models.Alert, error) {
	var user models.MongoUser
	err := coll.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}},
		options.FindOne().SetProjection(bson.D{primitive.E{Key: "_id", Value: 0},
			primitive.E{Key: "chat_id", Value: 0}, primitive.E{Key: "timestamp", Value: 0}})).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.Alerts, nil
}

func GetPairsByMarket(coll *mongo.Collection, id int64, market string, connected bool, ctx context.Context) ([]string, []models.Alert, error) {
	var user models.MongoUser
	err := coll.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&user)
	if err != nil {
		return nil, nil, err
	}
	if len(user.Alerts) == 0 {
		return nil, nil, NoPairsErr
	}

	pairs := make([]string, len(user.Alerts))
	alerts := make([]models.Alert, len(user.Alerts))

	index := 0
	for _, v := range user.Alerts {
		if v.Market == market && v.Connected == connected {
			pairs[index] = v.Pair
			alerts[index] = v
			index++
		}
	}

	if pairs[0] == "" {
		return nil, nil, NoPairsErr
	}

	pairs = pairs[:index]
	alerts = alerts[:index]

	fPairs := make([]string, index)
	fAlerts := make([]models.Alert, index)

	copy(fPairs, pairs)
	copy(fAlerts, alerts)

	return fPairs, fAlerts, nil
}

func splitPairs(result []interface{}) []string {
	if len(result) == 0 {
		return nil
	}
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
	pipeline := mongo.Pipeline{
		bson.D{
			primitive.E{
				Key: "$match", Value: bson.D{
					// primitive.E{Key: "updateDescription.updatedFields.alerts", Value: bson.D{primitive.E{Key: "$size", Value: 1}}},
					primitive.E{Key: "operationType", Value: "update"},
				},
			},
		},
	}
	cs, err := coll.Watch(ctx, pipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func WatchForChanges(coll *mongo.Collection, userUpdatePool map[int64]models.MongoUser, ctx context.Context) error {
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
		// fmt.Println(event["updateDescription"])
		fmt.Println(event)

		output, err := bson.Marshal(event["fullDocument"])
		if err != nil {
			return err
		}

		err = bson.Unmarshal(output, user)
		if err != nil {
			return err
		}
		userUpdatePool[user.UsedID] = *user
	}
	return nil
}

func CheckForChanges(id int64, market string, userUpdatePool map[int64]models.MongoUser, oldAlerts []models.Alert) []models.Alert {
	if len(userUpdatePool) == 0 {
		return nil
	}
	user, ok := userUpdatePool[id]
	if ok {
		return findMissingAlerts(user.Alerts, oldAlerts)
	}
	return nil
}

func findMissingAlerts(updated []models.Alert, old []models.Alert) []models.Alert {
	switch {
	case len(updated) > len(old):
		diff := len(updated) - len(old)
		pairs := make([]models.Alert, 0, diff)
		for i, v := range old {
			if old[i] != updated[i] {
				pairs = append(pairs, v)
			}
		}
		if len(pairs) < diff {
			diff = len(pairs) - diff
			pairs = append(pairs, pairs[len(pairs)-diff:]...)
		}
		return pairs
	case len(updated) < len(old):
		diff := len(old) - len(updated)
		pairs := make([]models.Alert, 0, diff)
		for i, v := range updated {
			if old[i] != updated[i] {
				pairs = append(pairs, v)
			}
		}
		if len(pairs) < diff {
			diff = len(pairs) - diff
			pairs = append(pairs, pairs[len(pairs)-diff:]...)
		}
		return pairs
	}
	return nil
}
