package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var GMongo *mongo.Client

func ChaosMongo(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	GMongo, err = mongo.Connect(ctx, options.Client().ApplyURI(dsn))
	if err != nil {
		return err
	}

	err = GMongo.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}

	return nil
}

func DisconnectMongo() {
	err := GMongo.Ping(context.Background(), readpref.Primary())
	if err == nil {
		err = GMongo.Disconnect(context.Background())
	}
}
