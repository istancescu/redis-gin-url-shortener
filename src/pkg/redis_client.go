package pkg

import (
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func CreateNewRedisClient(options *redis.Options) *RedisClient {
	client := RedisClient{Client: redis.NewClient(options)}
	return &client
}
