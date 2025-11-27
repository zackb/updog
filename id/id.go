package id

import (
	"crypto/rand"
	"log"
	"math/big"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const length = 11

func NewID() string {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			log.Println("Failed to generate ID:", err.Error())
			return ""
		}
		result[i] = alphabet[num.Int64()]
	}
	return string(result)
}
