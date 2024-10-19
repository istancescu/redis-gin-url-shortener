package pkg

import (
	"math/rand"
	"strconv"
)

func ShortenUrl() (string, string) {

	var randomIdAsString = GenerateRandomUrl()

	shortenedUrl := "http://localhost:8080/redirectTo/" + randomIdAsString

	return shortenedUrl, randomIdAsString
}

func GenerateRandomUrl() string {
	return strconv.FormatInt(9999999999-rand.Int63n(9000000000), 10)
}
