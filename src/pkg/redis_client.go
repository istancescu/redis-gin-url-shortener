package pkg

import (
	"github.com/redis/go-redis/v9"
	"time"
)

const (
	duration    = time.Hour * 25
	prefixHttp  = "http://"
	prefixHttps = "https://"
)

type RedisClient struct {
	client *redis.Client
}

func getRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     "localhost:6379",
		Password: "1234", // no password set
		DB:       0,      // use default DB
	}
}

func CreateNewRedisClient() *RedisClient {
	client := RedisClient{client: redis.NewClient(getRedisOptions())}
	return &client
}
