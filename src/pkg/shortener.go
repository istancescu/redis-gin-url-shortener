package pkg

import (
	"math/rand"
	"strconv"
)

func ShortenUrl() (string, string) {

	var randomIdAsString = generateRandomUrl()

	// TODO: this should be moved to config, use string literals
	shortenedUrl := "http://localhost:8080/redirectTo/" + randomIdAsString

	return shortenedUrl, randomIdAsString
}

// Lovely magic numbers no?
func generateRandomUrl() string {
	return strconv.FormatInt(9999999999-rand.Int63n(9000000000), 10)
}
