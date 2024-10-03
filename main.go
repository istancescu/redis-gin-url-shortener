package main

import (
	"awesomeProject/src/src/pkg"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	prefix      = "http://"
	prefixHttps = "https://"
)

type RedisClient struct {
	client *redis.Client
}

func main() {
	r := setupRouter()
	err := r.Run()

	if err != nil {
		return
	}
}

func getRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     "localhost:6379",
		Password: "1234", // no password set
		DB:       0,      // use default DB
	}

}

func setupRouter() *gin.Engine {
	r := gin.Default()

	client := redis.NewClient(getRedisOptions())

	redisClient := &RedisClient{client: client}

	r.GET("/url/:urlToShorten", redisClient.defaultPathHandler)
	r.GET("/redirectTo/:path", redisClient.redirectToHandler)

	return r
}

func (rc RedisClient) redirectToHandler(c *gin.Context) {
	ctx := c.Request.Context()

	redirectUrl := c.Param("path")

	foundKey, err := rc.client.Get(ctx, redirectUrl).Result()

	if errors.Is(err, redis.Nil) {
		log.Printf("Redirect url %s not exist", redirectUrl)
		c.AbortWithStatus(http.StatusNotFound)
	} else if err != nil {
		log.Printf("Redirect url %s not exist", redirectUrl)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := appendHttpsToUrl(foundKey)

	// Redirect to the dynamically constructed URL
	c.Redirect(http.StatusFound, url)

	return
}

func appendHttpsToUrl(foundKey string) string {
	var url string

	if urlContainsProtocolPrefix(foundKey) {
		url = foundKey
	} else {
		// If no protocol is specified, default to https
		url = prefixHttps + foundKey
	}
	return url
}

func urlContainsProtocolPrefix(foundKey string) bool {
	return strings.HasPrefix(foundKey, prefix) || strings.HasPrefix(foundKey, prefixHttps)
}

func (rc RedisClient) defaultPathHandler(c *gin.Context) {
	urlToShorten := c.Param("urlToShorten")

	ctx := c.Request.Context()

	_, err := rc.client.Get(ctx, urlToShorten).Result()

	if errors.Is(err, redis.Nil) {
		_, shortID := pkg.ShortenUrl()

		_, _ = rc.client.Set(ctx, urlToShorten, shortID, time.Hour*25).Result()
		_, _ = rc.client.Set(ctx, shortID, urlToShorten, time.Hour*25).Result()

		log.Printf("{%s} : {%s} \n", urlToShorten, shortID)
		log.Printf("{%s} : {%s} \n", shortID, urlToShorten)

	} else if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Printf("{%s} : {%s} \n", urlToShorten, err)
	}
	c.AbortWithStatus(http.StatusAlreadyReported)

	return
}
