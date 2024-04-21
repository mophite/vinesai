package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var GRedis *redis.Client

func ChaosRedis(address, password string) error {
	GRedis = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	return GRedis.Ping(context.Background()).Err()
}
