package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func CreateRedisClient(ConnectionString string) (*redis.Client, error) {
	opt, err := redis.ParseURL(ConnectionString)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	fmt.Println("Redis up and working ")
	return client, nil
}
