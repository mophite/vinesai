package db

import (
	"context"
	"vinesai/internel/ava"

	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var Mgo *mongo.Database

var mongoClient *mongo.Client

func ChaosMongo(dsn string) error {

	monitor := &event.CommandMonitor{
		Started: func(_ context.Context, evt *event.CommandStartedEvent) {
			ava.Debugf("Command: %s\nRequest: %+v\n", evt.CommandName, evt.Command)
		},
	}

	var err error
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(dsn).SetMonitor(monitor))
	if err != nil {
		return err
	}

	err = mongoClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}

	Mgo = mongoClient.Database("homingai")

	return nil
}

func DisconnectMongo() {
	err := mongoClient.Ping(context.Background(), readpref.Primary())
	if err == nil {
		err = mongoClient.Disconnect(context.Background())
	}
}
