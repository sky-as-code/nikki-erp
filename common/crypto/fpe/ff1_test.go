package fpe

import (
	"encoding/hex"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nistKey128 = "2B7E151628AED2A6ABF7158809CF4F3C"
	nistKey192 = "2B7E151628AED2A6ABF7158809CF4F3CEF4359D8D580AA4F"
	nistKey256 = "2B7E151628AED2A6ABF7158809CF4F3CEF4359D8D580AA4F7F036D6F04FC6A94"
	base36     = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func toNumerals(t *testing.T, text string) []uint16 {
	result := make([]uint16, len(text))
	for i, char := range text {
		index := strings.IndexRune(base36, char)
		require.GreaterOrEqual(t, index, 0)
		result[i] = uint16(index)
	}
	return result
}

func fromNumerals(numerals []uint16) string {
	var builder strings.Builder
	for _, numeral := range numerals {
		builder.WriteByte(base36[numeral])
	}
	return builder.String()
}

func mustHex(t *testing.T, text string) []byte {
	decoded, err := hex.DecodeString(text)
	require.NoError(t, err)
	return decoded
}

// NIST SP 800-38G FF1 samples 1-9 (FF1samples.pdf).
func TestFf1NistSamples(t *testing.T) {
	cases := []struct {
		name, key, tweak, plain, cipher string
		radix                           int
	}{
		{"sample1", nistKey128, "", "0123456789", "2433477484", 10},
		{"sample2", nistKey128, "39383736353433323130", "0123456789", "6124200773", 10},
		{"sample3", nistKey128, "3737373770717273373737", "0123456789abcdefghi", "a9tv40mll9kdu509eum", 36},
		{"sample4", nistKey192, "", "0123456789", "2830668132", 10},
		{"sample5", nistKey192, "39383736353433323130", "0123456789", "2496655549", 10},
		{"sample6", nistKey192, "3737373770717273373737", "0123456789abcdefghi", "xbj3kv35jrawxv32ysr", 36},
		{"sample7", nistKey256, "", "0123456789", "6657667009", 10},
		{"sample8", nistKey256, "39383736353433323130", "0123456789", "1001623463", 10},
		{"sample9", nistKey256, "3737373770717273373737", "0123456789abcdefghi", "xs8a0azh2avyalyzuwd", 36},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ff1, err := NewFf1(mustHex(t, tc.key), tc.radix)
			require.NoError(t, err)
			tweak := mustHex(t, tc.tweak)
			encrypted, err := ff1.Encrypt(tweak, toNumerals(t, tc.plain))
			require.NoError(t, err)
			assert.Equal(t, tc.cipher, fromNumerals(encrypted))
			decrypted, err := ff1.Decrypt(tweak, encrypted)
			require.NoError(t, err)
			assert.Equal(t, tc.plain, fromNumerals(decrypted))
		})
	}
}

func randomNumerals(rng *rand.Rand, length, radix int) []uint16 {
	result := make([]uint16, length)
	for i := range result {
		result[i] = uint16(rng.Intn(radix))
	}
	return result
}

func TestFf1Radix32RoundTrip(t *testing.T) {
	ff1, err := NewFf1(mustHex(t, nistKey256), 32)
	require.NoError(t, err)
	rng := rand.New(rand.NewSource(42))
	tweak := mustHex(t, "000102030405060708090a0b0c0d0e0f")
	for length := 4; length <= 12; length++ {
		for iter := 0; iter < 200; iter++ {
			plain := randomNumerals(rng, length, 32)
			encrypted, err := ff1.Encrypt(tweak, plain)
			require.NoError(t, err)
			require.Len(t, encrypted, length)
			for _, numeral := range encrypted {
				require.Less(t, int(numeral), 32)
			}
			decrypted, err := ff1.Decrypt(tweak, encrypted)
			require.NoError(t, err)
			require.Equal(t, plain, decrypted)
		}
	}
}

func TestFf1NonIdentityAndTweakSensitivity(t *testing.T) {
	ff1, err := NewFf1(mustHex(t, nistKey256), 32)
	require.NoError(t, err)
	plain := make([]uint16, 6)
	identical, sameAcrossTweaks := 0, 0
	for value := 0; value < 500; value++ {
		plain[5], plain[4] = uint16(value%32), uint16(value/32)
		first, err := ff1.Encrypt([]byte("tweak-one"), plain)
		require.NoError(t, err)
		second, err := ff1.Encrypt([]byte("tweak-two"), plain)
		require.NoError(t, err)
		if fromNumerals(first) == fromNumerals(plain) {
			identical++
		}
		if fromNumerals(first) == fromNumerals(second) {
			sameAcrossTweaks++
		}
	}
	assert.LessOrEqual(t, identical, 1)
	assert.LessOrEqual(t, sameAcrossTweaks, 1)
}

func TestFf1OtherRadixRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	cases := []struct{ radix, length int }{{2, 20}, {2, 64}, {10, 6}, {256, 8}, {65536, 2}, {65536, 12}}
	for _, tc := range cases {
		ff1, err := NewFf1(mustHex(t, nistKey192), tc.radix)
		require.NoError(t, err)
		plain := randomNumerals(rng, tc.length, tc.radix)
		tweak := []byte(strings.Repeat("x", 37))
		encrypted, err := ff1.Encrypt(tweak, plain)
		require.NoError(t, err)
		decrypted, err := ff1.Decrypt(tweak, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plain, decrypted, "radix %d", tc.radix)
	}
}

func TestFf1InvalidInput(t *testing.T) {
	_, err := NewFf1(make([]byte, 15), 10)
	assert.Error(t, err)
	_, err = NewFf1(make([]byte, 16), 1)
	assert.Error(t, err)
	_, err = NewFf1(make([]byte, 16), 65537)
	assert.Error(t, err)

	ff1, err := NewFf1(make([]byte, 32), 10)
	require.NoError(t, err)
	_, err = ff1.Encrypt(nil, []uint16{1, 2, 3, 4, 5})
	assert.Error(t, err, "domain below one million")
	_, err = ff1.Encrypt(nil, []uint16{1, 2, 3, 4, 5, 10})
	assert.Error(t, err, "numeral not below radix")
	_, err = ff1.Encrypt(nil, make([]uint16, 57))
	assert.Error(t, err, "longer than maxlen")
}
