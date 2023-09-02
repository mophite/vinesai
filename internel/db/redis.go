package db

import (
	"context"
	"github.com/redis/go-redis/v9"
)

var GRedis *redis.Client

func ChaosRedis(address string) error {
	GRedis = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		//Password: "kadCPpsEXbxtzcKD", // no password set
		DB: 0, // use default DB
	})

	return GRedis.Ping(context.Background()).Err()
}
