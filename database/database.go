package database

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func CreateClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:30073",
		Password: "mypassword",
		DB:       0,
	})
	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		panic(err)
	}

	return rdb
}
