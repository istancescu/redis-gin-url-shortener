package main

import (
	"awesomeProject/src/config"
	"awesomeProject/src/pkg"
	"github.com/gin-gonic/gin"
)

const (
	configFilePath string = "config.yaml"
)

func main() {
	redisConfig := config.ProvideRedisConfig(configFilePath)
	client := pkg.CreateNewRedisClient(redisConfig)
	router := setupRouter(client)

	err := router.Run()

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
