package crypto

import (
	"crypto/sha512"
	"encoding/base64"
)

// Combine text and salt then hash them using the SHA-512
// hashing algorithm and then return the hashed text
// as a base64 encoded string
func Hash(text string, salt string) string {
	// Convert text string to byte slice
	var textBytes = []byte(text)
	var saltBytes = []byte(salt)

	// Create sha-512 hasher
	var sha512Hasher = sha512.New()

	// Append salt to text
	textBytes = append(textBytes, saltBytes...)

	// Write text bytes to the hasher
	sha512Hasher.Write(textBytes)

	// Get the SHA-512 hashed text
	var hashedTextBytes = sha512Hasher.Sum(nil)

	// Convert the hashed text to a base64 encoded string
	var base64EncodedtextHash = base64.URLEncoding.EncodeToString(hashedTextBytes)

	return base64EncodedtextHash
}
