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

	return
}

func (rc *RedisClient) DefaultPathHandler(c *gin.Context) {
	//TODO simplify this logic
	urlToShorten := c.Param("urlToShorten")

	ctx := c.Request.Context()

	// urlToShorten (google.com) is used to retrieve key if already existent (1234)
	foundUrl, err := rc.Client.Get(ctx, urlToShorten).Result()

	if errors.Is(err, redis.Nil) {
		shortID := rc.addCreatedUrlToRedis(ctx, urlToShorten)
		c.JSONP(http.StatusCreated, gin.H{
			"message":      "Url created successfully",
			"shortenedUrl": shortID,
		})

		return
	} else if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Printf("{%s} : {%s} \n", urlToShorten, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "URL already cached",
		"shortenedUrl": foundUrl})

	return
}

func (rc *RedisClient) addCreatedUrlToRedis(ctx context.Context, urlToShorten string) string {
	_, shortID := ShortenUrl()

	//TODO error handling
	_, _ = rc.Client.Set(ctx, urlToShorten, shortID, duration).Result()
	_, _ = rc.Client.Set(ctx, shortID, urlToShorten, duration).Result()

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
