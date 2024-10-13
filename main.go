package main

import (
	"awesomeProject/src/pkg"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func getRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     "localhost:6379",
		Password: "1234", // no password set
		DB:       0,      // use default DB
	}
}

func main() {
	client := pkg.CreateNewRedisClient(getRedisOptions())
	r := setupRouter(client)

	err := r.Run()

	if err != nil {
		return
	}
}

func setupRouter(client *pkg.RedisClient) *gin.Engine {
	r := gin.Default()

	r.GET("/url/:urlToShorten", pkg.DefaultPathHandler(client))
	r.GET("/redirectTo/:path", pkg.RedirectToHandler(client))

	return r
}
