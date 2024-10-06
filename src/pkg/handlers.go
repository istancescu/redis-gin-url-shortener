package pkg

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	duration    = time.Hour * 25
	prefixHttp  = "http://"
	prefixHttps = "https://"
)

type KeyValueStore interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
}

func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return rc.Client.Get(ctx, key).Result()
}

// Set stores a key-value pair in Redis
func (rc *RedisClient) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	_, err := rc.Client.Set(ctx, key, value, expiration).Result()
	return err
}

func (rc *RedisClient) RedirectToHandler(c *gin.Context) {
	ctx := c.Request.Context()

	redirectUrl := c.Param("path")

	foundKey, err := rc.Client.Get(ctx, redirectUrl).Result()

	if err != nil {
		log.Printf("Redirect url %s not exist / other error occured", redirectUrl)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := AppendHttpsToUrl(foundKey)

	// Redirect to the dynamically constructed URL
	c.Redirect(http.StatusFound, url)
}

func (rc *RedisClient) DefaultPathHandler(c *gin.Context) {
	//TODO simplify this logic
	urlToShorten := c.Param("urlToShorten")

	ctx := c.Request.Context()

	// urlToShorten (google.com) is used to retrieve key if already existent (1234)
	foundUrl, err := rc.Get(ctx, urlToShorten)

	if errors.Is(err, redis.Nil) {
		shortID := rc.createShortenedUrl(ctx, urlToShorten)
		c.JSONP(http.StatusCreated, gin.H{
			"message":      "Url created successfully",
			"shortenedUrl": shortID,
		})
		return
	}

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Printf("{%s} : {%s} \n", urlToShorten, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "URL already cached",
		"shortenedUrl": foundUrl})
}

func (rc *RedisClient) createShortenedUrl(ctx context.Context, urlToShorten string) string {
	_, shortID := ShortenUrl()

	//TODO error handling
	_ = rc.Client.Set(ctx, urlToShorten, shortID, duration)
	_ = rc.Client.Set(ctx, shortID, urlToShorten, duration)

	log.Printf("{%s} : {%s} \n", urlToShorten, shortID)
	log.Printf("{%s} : {%s} \n", shortID, urlToShorten)
	return shortID
}

func AppendHttpsToUrl(foundKey string) string {
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
	return strings.HasPrefix(foundKey, prefixHttp) || strings.HasPrefix(foundKey, prefixHttps)
}
