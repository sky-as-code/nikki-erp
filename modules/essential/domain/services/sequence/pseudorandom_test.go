package sequence

import (
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/crypto/fpe"
	"github.com/sky-as-code/nikki-erp/common/encoding/crockford"
)

const (
	testKeyHex    = "2b7e151628aed2a6abf7158809cf4f3cef4359d8d580aa4f7f036d6f04fc6a94"
	testTweakHex  = "000102030405060708090a0b0c0d0e0f"
	otherTweakHex = "f0e0d0c0b0a090807060504030201000"
)

func TestLengthDomainBoundaries(t *testing.T) {
	cases := []struct {
		counter    int64
		wantLength int
		wantOffset uint64
	}{
		{1, 6, 0},
		{1_073_741_824, 6, 1_073_741_823},
		{1_073_741_825, 7, 0},
		{35_433_480_192, 7, 34_359_738_367},
		{35_433_480_193, 8, 0},
		{1_134_945_107_968, 8, 1_099_511_627_775},
		{1_134_945_107_969, 9, 0},
		{36_319_317_196_800, 9, 35_184_372_088_831},
	}
	for _, tc := range cases {
		offset, length, err := lengthDomain(tc.counter, 1)
		require.NoError(t, err, tc.counter)
		assert.Equal(t, tc.wantLength, length, tc.counter)
		assert.Equal(t, tc.wantOffset, offset, tc.counter)

		value, err := Pseudorandom(tc.counter, 1, testKeyHex, testTweakHex)
		require.NoError(t, err, tc.counter)
		assert.Len(t, value, tc.wantLength, tc.counter)
		assertCrockford(t, value)
	}
}

func TestPseudorandomCapacityExceeded(t *testing.T) {
	_, err := Pseudorandom(36_319_317_196_801, 1, testKeyHex, testTweakHex)
	assert.ErrorIs(t, err, ErrCapacityExceeded)
	assert.Equal(t, CodeCapacityExceeded, err.Error())

	_, err = Pseudorandom(36_319_317_196_809, 10, testKeyHex, testTweakHex)
	require.NoError(t, err, "initial value shifts the capacity window")
	_, err = Pseudorandom(36_319_317_196_810, 10, testKeyHex, testTweakHex)
	assert.ErrorIs(t, err, ErrCapacityExceeded)
}

func TestPseudorandomInvalidInput(t *testing.T) {
	_, err := Pseudorandom(0, 1, testKeyHex, testTweakHex)
	assert.Error(t, err)
	_, err = Pseudorandom(1, 1, "zz", testTweakHex)
	assert.Error(t, err)
	_, err = Pseudorandom(1, 1, testKeyHex[:32], testTweakHex)
	assert.Error(t, err, "AES-128 key is refused: AES-256 only")
	_, err = Pseudorandom(1, 1, testKeyHex, testTweakHex[:30])
	assert.Error(t, err)
	assert.False(t, errors.Is(err, ErrCapacityExceeded))
}

func assertCrockford(t *testing.T, value string) {
	for _, char := range value {
		require.True(t, strings.ContainsRune(crockford.Alphabet, char), "char %q in %s", char, value)
	}
}

// Within one length domain distinct counters give distinct values, and decrypting gives the offset back.
func TestPseudorandomBijective(t *testing.T) {
	key, _ := hex.DecodeString(testKeyHex)
	tweak, _ := hex.DecodeString(testTweakHex)
	ff1, err := fpe.NewFf1(key, crockford.Radix)
	require.NoError(t, err)
	domainStarts := []int64{1, 1_073_741_825, 35_433_480_193, 1_134_945_107_969}
	for index, start := range domainStarts {
		seen := map[string]int64{}
		for step := int64(0); step < 3000; step++ {
			counter := start + step*7919
			value, err := Pseudorandom(counter, 1, testKeyHex, testTweakHex)
			require.NoError(t, err)
			require.Len(t, value, MinLength+index)
			previous, duplicate := seen[value]
			require.False(t, duplicate, "counters %d and %d both gave %s", previous, counter, value)
			seen[value] = counter

			numerals, err := crockford.ToNumerals(value)
			require.NoError(t, err)
			plain, err := ff1.Decrypt(tweak, numerals)
			require.NoError(t, err)
			offset, err := crockford.Decode(crockford.FromNumerals(plain))
			require.NoError(t, err)
			require.Equal(t, uint64(counter-start), offset)
		}
	}
}

func TestPseudorandomTweakChangesOutput(t *testing.T) {
	differs := 0
	for counter := int64(1); counter <= 200; counter++ {
		first, err := Pseudorandom(counter, 1, testKeyHex, testTweakHex)
		require.NoError(t, err)
		again, err := Pseudorandom(counter, 1, testKeyHex, testTweakHex)
		require.NoError(t, err)
		require.Equal(t, first, again, "deterministic")
		second, err := Pseudorandom(counter, 1, testKeyHex, otherTweakHex)
		require.NoError(t, err)
		if first != second {
			differs++
		}
	}
	assert.GreaterOrEqual(t, differs, 199)
}

func TestNewSecrets(t *testing.T) {
	key, err := NewKeyHex()
	require.NoError(t, err)
	assert.Len(t, key, 64)
	tweak, err := NewTweakHex()
	require.NoError(t, err)
	assert.Len(t, tweak, 32)
	otherKey, err := NewKeyHex()
	require.NoError(t, err)
	assert.NotEqual(t, key, otherKey)

	value, err := Pseudorandom(42, 1, key, tweak)
	require.NoError(t, err)
	assert.Len(t, value, MinLength)
}
