package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Mgo *mongo.Database

var mongoClient *mongo.Client

func ChaosMongo(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))
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
