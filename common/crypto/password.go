package crypto

import (
	"crypto/rand"
	"math/big"

	"go.bryk.io/pkg/errors"
)

var (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars  = "0123456789"
	specialChars = "!@#$%^&*()-_=+[]{}<>?/|"
	allChars     = lowerChars + upperChars + numberChars + specialChars
)

func GenerateSecurePassword(length int) (string, error) {
	if length < 4 {
		return "", errors.Errorf("password length must be at least 4")
	}

	var password []byte

	// Ensure at least one character from each required set
	requiredSets := []string{lowerChars, upperChars, numberChars, specialChars}
	for _, set := range requiredSets {
		c, err := randomChar(set)
		if err != nil {
			return "", err
		}
		password = append(password, c)
	}

	// Fill the remaining length with random characters from all sets
	for i := len(password); i < length; i++ {
		c, err := randomChar(allChars)
		if err != nil {
			return "", err
		}
		password = append(password, c)
	}

	// Shuffle to avoid predictable character positions
	shuffled, err := shuffleBytes(password)
	if err != nil {
		return "", err
	}

	return string(shuffled), nil
}

func randomChar(charset string) (byte, error) {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[i.Int64()], nil
}

func shuffleBytes(b []byte) ([]byte, error) {
	n := len(b)
	for i := n - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := int(jBig.Int64())
		b[i], b[j] = b[j], b[i]
	}
	return b, nil
}
