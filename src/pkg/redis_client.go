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
	duration    = time.Hour * 24
	prefixHttp  = "http://"
	prefixHttps = "https://"
)

type RedisClient struct {
	Client *redis.Client
}

func CreateNewRedisClient(options *redis.Options) *RedisClient {
	client := RedisClient{Client: redis.NewClient(options)}
	return &client
}

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

func RedirectToHandler(store KeyValueStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		redirectUrl := c.Param("path")

		foundKey, err := store.Get(ctx, redirectUrl)

		if err != nil {
			log.Printf("Redirect url %s not exist / other error occured", redirectUrl)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		url := appendHttpsToUrl(foundKey)

		// Redirect to the dynamically constructed URL
		c.Redirect(http.StatusFound, url)
	}
}

func DefaultPathHandler(store KeyValueStore) gin.HandlerFunc {
	//TODO simplify this logic
	return func(c *gin.Context) {
		urlToShorten := c.Param("urlToShorten")

		ctx := c.Request.Context()

		// urlToShorten (google.com) is used to retrieve key if already existent (1234)
		foundUrl, err := store.Get(ctx, urlToShorten)

		if errors.Is(err, redis.Nil) {
			shortID := createShortenedUrl(store, ctx, urlToShorten)
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
}

func createShortenedUrl(store KeyValueStore, ctx context.Context, urlToShorten string) string {
	_, shortID := ShortenUrl()

	//TODO error handling
	_ = store.Set(ctx, urlToShorten, shortID, duration)
	_ = store.Set(ctx, shortID, urlToShorten, duration)

	log.Printf("{%s} : {%s} \n", urlToShorten, shortID)
	log.Printf("{%s} : {%s} \n", shortID, urlToShorten)
	return shortID
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
	return strings.HasPrefix(foundKey, prefixHttp) || strings.HasPrefix(foundKey, prefixHttps)
}
