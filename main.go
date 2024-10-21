package main

import (
	"awesomeProject/src/alb"
	"awesomeProject/src/config"
	"awesomeProject/src/pkg"
	"fmt"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"log"
	"net/url"
	"sync"
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

	createAndRunLoadBalancer()

	spawnMultipleServers(router, 3, []uint16{8080, 8081, 8083})

}

// TODO refactor this
func runLoadBalancer(router *gin.Engine) {
	go func() {
		albConfig := alb.ServerConfiguration{Timeout: 15}

		alb1Url, _ := url.Parse("http://localhost:8080")
		alb2Url, _ := url.Parse("http://localhost:8081")

		alb1 := alb.CreateAppServer(albConfig, alb1Url)
		alb2 := alb.CreateAppServer(albConfig, alb2Url)
		alb1.SetAlive(true)
		alb2.SetAlive(true)

		sp := alb.CreateServerPool()
		alb.AddServer(sp, alb1)
		alb.AddServer(sp, alb2)

		alb.CreateLoadBalancer(sp)

		var port = 9000

		router.GET("/api/router/:urlToShorten", func(c *gin.Context) {
			// Use the server pool to handle the incoming request and forward it
			log.Printf("running alb on port: %d \n", port)
			urlToShorten := c.Param("urlToShorten")
			log.Printf("Received request to shorten: %s \n", urlToShorten)

			sp.HandleHTTPRequests(c)
		})

		_ = router.Run(fmt.Sprintf(":%d", port))
	}()
}

func createAndRunLoadBalancer() *gin.Engine {
	// Initialize the Gin router
	router := gin.Default()

	// Apply any middlewares, like CORS
	router.Use(cors.Default())

	// Call the function to run the load balancer
	runLoadBalancer(router)

	return router
}

func spawnMultipleServers(router *gin.Engine, count uint8, port []uint16) {
	var wg sync.WaitGroup
	wg.Add(int(count))
	for i := 0; i < int(count); i++ {
		go func(i int) {
			defer wg.Done()
			err := router.Run(fmt.Sprintf(":%d", port[i]))
			if err != nil {
				panic(err)
			}
		}(i)
	}
	wg.Wait()
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
