package main

import (
	"awesomeProject/src/pkg"
	"github.com/gin-gonic/gin"
)

func main() {
	client := pkg.CreateNewRedisClient()
	r := setupRouter(client)

	err := r.Run()

	if err != nil {
		return
	}
}

func setupRouter(client *pkg.RedisClient) *gin.Engine {
	r := gin.Default()

	r.GET("/url/:urlToShorten", client.DefaultPathHandler)
	r.GET("/redirectTo/:path", client.RedirectToHandler)

	return r
}
