package utils

import (
	"crypto/rand"
)

/*
GenerateRandomStr generates a random string.
*/
func GenerateRandomStr() string {
	return rand.Text()
}

/*
Truncate truncates a string to the specified length.
*/
func Truncate(s string, length int) string {
	if length < len(s) {
		return s[:length]
	}

	return s
}
