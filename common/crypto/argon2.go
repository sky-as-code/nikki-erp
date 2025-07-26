package crypto

import (
	"crypto/rand"
	"encoding/base64"
	stdErr "errors"
	"fmt"
	"strconv"
	"strings"

	"go.bryk.io/pkg/errors"
	"golang.org/x/crypto/argon2"
)

const (
	saltLength = 16
	memory     = 64 * 1024 // 64 MB
	iterations = 3         // iterations
	threads    = 2
	keyLength  = 32
	version    = argon2.Version
)

// Format: $argon2id${version}${memory}${iterations}${threads}${salt_base64}${hash_base64}
func GenerateFromPassword(password []byte) ([]byte, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, errors.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(password, salt, iterations, memory, threads, keyLength)

	encoded := fmt.Sprintf("$argon2id$%d$%d$%d$%d$%s$%s",
		version,
		memory,
		iterations,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return []byte(encoded), nil
}

func CompareHashAndPassword(encodedHash []byte, password []byte) (bool, error) {
	parts := strings.Split(string(encodedHash), "$")
	if len(parts) != 8 {
		return false, errors.Errorf("invalid hash format")
	}

	// parts[0] is empty due to leading '$'
	// algo := parts[1] // "argon2id"
	// ver := parts[2]    // "19"
	var err error
	num, nerr := strconv.ParseUint(parts[3], 10, 32)
	err = stdErr.Join(err, nerr)
	mem := uint32(num) // "65536"

	num, err = strconv.ParseUint(parts[4], 10, 32)
	err = stdErr.Join(err, nerr)
	iter := uint32(num) // "3"

	num, err = strconv.ParseUint(parts[5], 10, 8)
	err = stdErr.Join(err, nerr)
	thr := uint8(num) // "2"

	salt64 := parts[6] // base64 salt
	hash64 := parts[7] // base64 hash

	if err != nil {
		return false, errors.Errorf("invalid hash format: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(salt64)
	if err != nil {
		return false, errors.Errorf("invalid salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hash64)
	if err != nil {
		return false, errors.Errorf("invalid hash: %w", err)
	}

	// Rehash using same parameters
	hash := argon2.IDKey(password, salt, iter, mem, thr, uint32(len(expectedHash)))

	if subtleCompare(hash, expectedHash) {
		return true, nil
	}
	return false, nil
}

func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var res byte
	for i := 0; i < len(a); i++ {
		res |= a[i] ^ b[i]
	}
	return res == 0
}
