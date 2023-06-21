package helpers

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenerateUniqueShortString(length int) string {
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}

	return string(result)
}
