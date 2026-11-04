// Package crockford implements fixed-width Crockford Base32 with integer arithmetic only.
package crockford

import (
	"strings"

	"go.bryk.io/pkg/errors"
)

const (
	Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	Radix    = 32

	bitsPerChar = 5
	// maxWidth is the widest string whose value can still fit in a uint64 (13*5 = 65 bits, top bit unused).
	maxWidth = 13
)

var decodeTable = buildDecodeTable()

func buildDecodeTable() [256]int8 {
	var table [256]int8
	for i := range table {
		table[i] = -1
	}
	for i := 0; i < len(Alphabet); i++ {
		table[Alphabet[i]] = int8(i)
		table[strings.ToLower(Alphabet[i : i+1])[0]] = int8(i)
	}
	for _, alias := range []struct {
		char  byte
		value int8
	}{{'I', 1}, {'i', 1}, {'L', 1}, {'l', 1}, {'O', 0}, {'o', 0}} {
		table[alias.char] = alias.value
	}
	return table
}

// EncodeFixed encodes n as exactly width characters, left-padded with '0'.
// It fails when n needs more than width characters.
func EncodeFixed(n uint64, width int) (string, error) {
	if width <= 0 {
		return "", errors.Errorf("crockford: width %d must be positive", width)
	}
	if width < maxWidth && n>>(uint(width)*bitsPerChar) != 0 {
		return "", errors.Errorf("crockford: value %d does not fit in %d characters", n, width)
	}
	output := make([]byte, width)
	for i := width - 1; i >= 0; i-- {
		output[i] = Alphabet[n&(Radix-1)]
		n >>= bitsPerChar
	}
	return string(output), nil
}

// Decode parses a Crockford Base32 string, case-insensitively and accepting I/L as 1 and O as 0.
func Decode(text string) (uint64, error) {
	numerals, err := ToNumerals(text)
	if err != nil {
		return 0, err
	}
	var result uint64
	for _, numeral := range numerals {
		if result>>(64-bitsPerChar) != 0 {
			return 0, errors.Errorf("crockford: %q overflows uint64", text)
		}
		result = result<<bitsPerChar | uint64(numeral)
	}
	return result, nil
}

// ToNumerals maps each character to its value 0..31, with the same aliases as Decode.
func ToNumerals(text string) ([]uint16, error) {
	if text == "" {
		return nil, errors.New("crockford: empty input")
	}
	numerals := make([]uint16, len(text))
	for i := 0; i < len(text); i++ {
		value := decodeTable[text[i]]
		if value < 0 {
			return nil, errors.Errorf("crockford: invalid character %q at index %d", text[i], i)
		}
		numerals[i] = uint16(value)
	}
	return numerals, nil
}

// FromNumerals maps values 0..31 to canonical upper-case characters. Values must already be below 32.
func FromNumerals(numerals []uint16) string {
	output := make([]byte, len(numerals))
	for i, numeral := range numerals {
		output[i] = Alphabet[numeral&(Radix-1)]
	}
	return string(output)
}
