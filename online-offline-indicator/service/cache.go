package service

import (
	"fmt"

	"github.com/go-redis/redis"
)

func NewRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if _, err := client.Ping().Result(); err != nil {
		return nil, fmt.Errorf("failed to ping. [%w]", err)
	}
	return client, nil
}
