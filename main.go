package main

import (
	"awesomeProject/src/config"
	"awesomeProject/src/pkg"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"log"
)

const (
	configFilePath string = "config.yaml"
)

func main() {
	redisConfig, err := config.ProvideRedisConfig(configFilePath)
	if err != nil {
		log.Panicf("Error reading from yaml")
	}
	client := pkg.CreateNewRedisClient(redisConfig)
	router := setupRouter(client)

	err = router.Run()

	if err != nil {
		return
	}
}

func setupRouter(client *pkg.RedisClient) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	}))

	r.GET("/url/:urlToShorten", pkg.DefaultPathHandler(client))
	r.GET("/redirectTo/:path", pkg.RedirectToHandler(client))

	return r
}
