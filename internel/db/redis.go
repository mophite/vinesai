package db

import (
	"context"
	"vinesai/internel/ava"
	"vinesai/internel/x"

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

func RedisGet(key string, v interface{}) error {
	result, err := GRedis.Get(context.Background(), key).Result()
	if err != nil {
		return err
	}

	err = x.MustNativeUnmarshal([]byte(result), v)
	return err
}

func RedisSet(key string, value interface{}) error {
	return GRedis.Set(context.Background(), key, x.MustMarshal2String(value), 0).Err()
}
