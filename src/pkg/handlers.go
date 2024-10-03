package pkg

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strings"
)

func (rc *RedisClient) RedirectToHandler(c *gin.Context) {
	ctx := c.Request.Context()

	redirectUrl := c.Param("path")

	foundKey, err := rc.client.Get(ctx, redirectUrl).Result()

	if errors.Is(err, redis.Nil) {
		log.Printf("Redirect url %s not exist", redirectUrl)
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Redirect url %s not exist", redirectUrl)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := AppendHttpsToUrl(foundKey)

	// Redirect to the dynamically constructed URL
	c.Redirect(http.StatusFound, url)

	return
}

func (rc *RedisClient) DefaultPathHandler(c *gin.Context) {
	urlToShorten := c.Param("urlToShorten")

	ctx := c.Request.Context()

	_, err := rc.client.Get(ctx, urlToShorten).Result()

	if errors.Is(err, redis.Nil) {
		_, shortID := ShortenUrl()

		_, _ = rc.client.Set(ctx, urlToShorten, shortID, duration).Result()
		_, _ = rc.client.Set(ctx, shortID, urlToShorten, duration).Result()

		log.Printf("{%s} : {%s} \n", urlToShorten, shortID)
		log.Printf("{%s} : {%s} \n", shortID, urlToShorten)

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

	foundUrl, err := rc.client.Get(ctx, urlToShorten).Result()

	if err != nil {
		log.Printf("{%s} : {%s} \n", urlToShorten, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "URL already cached",
		"shortenedUrl": foundUrl})

	return
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
