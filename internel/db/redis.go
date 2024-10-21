package db

import (
	"context"
	"vinesai/internel/ava"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

var GRedis *redis.Client

var RedisLock *redsync.Redsync

func ChaosRedis(address, password string) error {
	GRedis = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	err := GRedis.Ping(context.Background()).Err()
	if err != nil {
		ava.Error(err)
		return err
	}

	pool := goredis.NewPool(GRedis)
	RedisLock = redsync.New(pool)

	return nil
}
