package util

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomStr(charset string, length int) (string, error) {
	result := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}

		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
