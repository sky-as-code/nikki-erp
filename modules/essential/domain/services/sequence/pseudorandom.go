package sequence

import (
	"crypto/rand"
	"encoding/hex"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/crypto/fpe"
	"github.com/sky-as-code/nikki-erp/common/encoding/crockford"
)

const (
	// MinLength and MaxLength bound the pseudorandom value length (requirement §3.1); they are not configurable.
	MinLength = 6
	MaxLength = 9

	KeyBytes   = 32
	TweakBytes = 16

	bitsPerChar = 5
)

// Pseudorandom turns counter value n into its public reference: the offset n-initialValue is placed in the
// shortest length domain (6..9 Crockford Base32 characters) that still has room, then FF1-encrypted with radix 32.
// Each length is its own FF1 domain, so values of different lengths never collide.
func Pseudorandom(n, initialValue int64, keyHex, tweakHex string) (string, error) {
	offset, length, err := lengthDomain(n, initialValue)
	if err != nil {
		return "", err
	}
	key, err := decodeHexSecret(keyHex, KeyBytes, "key")
	if err != nil {
		return "", err
	}
	tweak, err := decodeHexSecret(tweakHex, TweakBytes, "tweak")
	if err != nil {
		return "", err
	}
	plaintext, err := crockford.EncodeFixed(offset, length)
	if err != nil {
		return "", err
	}
	numerals, err := crockford.ToNumerals(plaintext)
	if err != nil {
		return "", err
	}
	cipher, err := fpe.NewFf1(key, crockford.Radix)
	if err != nil {
		return "", err
	}
	encrypted, err := cipher.Encrypt(tweak, numerals)
	if err != nil {
		return "", err
	}
	return crockford.FromNumerals(encrypted), nil
}

// lengthDomain walks the length domains with integer arithmetic: 32^length == 1 << (5*length).
func lengthDomain(n, initialValue int64) (uint64, int, error) {
	if n < initialValue {
		return 0, 0, errors.Errorf("sequence: counter value %d is below initial value %d", n, initialValue)
	}
	offset := uint64(n) - uint64(initialValue)
	length := MinLength
	for length <= MaxLength && offset >= uint64(1)<<(bitsPerChar*length) {
		offset -= uint64(1) << (bitsPerChar * length)
		length++
	}
	if length > MaxLength {
		return 0, 0, ErrCapacityExceeded
	}
	return offset, length, nil
}

func decodeHexSecret(text string, size int, name string) ([]byte, error) {
	decoded, err := hex.DecodeString(text)
	if err != nil {
		return nil, errors.Wrapf(err, "sequence: %s is not hex", name)
	}
	if len(decoded) != size {
		return nil, errors.Errorf("sequence: %s must be %d bytes, got %d", name, size, len(decoded))
	}
	return decoded, nil
}

// NewKeyHex returns a fresh 32-byte AES-256 key from crypto/rand, hex-encoded.
func NewKeyHex() (string, error) {
	return randomHex(KeyBytes)
}

// NewTweakHex returns a fresh 16-byte FF1 tweak from crypto/rand, hex-encoded.
func NewTweakHex() (string, error) {
	return randomHex(TweakBytes)
}

func randomHex(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", errors.Wrap(err, "sequence: read crypto/rand")
	}
	return hex.EncodeToString(buffer), nil
}
