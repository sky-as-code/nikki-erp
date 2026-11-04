package crockford

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeFixed(t *testing.T) {
	cases := []struct {
		value uint64
		width int
		want  string
	}{
		{0, 6, "000000"},
		{1, 6, "000001"},
		{31, 1, "Z"},
		{32, 2, "10"},
		{1073741823, 6, "ZZZZZZ"},
		{1073741824, 7, "1000000"},
		{math.MaxUint64, 13, "FZZZZZZZZZZZZ"},
		{5, 20, "00000000000000000005"},
	}
	for _, tc := range cases {
		got, err := EncodeFixed(tc.value, tc.width)
		require.NoError(t, err)
		assert.Equal(t, tc.want, got)
		decoded, err := Decode(got)
		require.NoError(t, err)
		assert.Equal(t, tc.value, decoded)
	}
}

func TestEncodeFixedOverflow(t *testing.T) {
	_, err := EncodeFixed(1073741824, 6)
	assert.Error(t, err)
	_, err = EncodeFixed(32, 1)
	assert.Error(t, err)
	_, err = EncodeFixed(1, 0)
	assert.Error(t, err)
}

func TestDecodeAliasesAndCase(t *testing.T) {
	cases := map[string]uint64{"zz": 1023, "Ii": 33, "lL": 33, "oO": 0, "1o": 32, "abc": 10*1024 + 11*32 + 12}
	for input, want := range cases {
		got, err := Decode(input)
		require.NoError(t, err, input)
		assert.Equal(t, want, got, input)
	}
}

func TestDecodeErrors(t *testing.T) {
	for _, input := range []string{"", "U", "AB-C", "ZZZZZZZZZZZZZZ", "GZZZZZZZZZZZZ"} {
		_, err := Decode(input)
		assert.Error(t, err, input)
	}
	_, err := Decode("FZZZZZZZZZZZZ")
	assert.NoError(t, err)
}

func TestNumerals(t *testing.T) {
	numerals, err := ToNumerals("0aZ-")
	assert.Error(t, err)
	assert.Nil(t, numerals)

	numerals, err = ToNumerals("09AhjZ")
	require.NoError(t, err)
	assert.Equal(t, []uint16{0, 9, 10, 17, 18, 31}, numerals)
	assert.Equal(t, "09AHJZ", FromNumerals(numerals))
	assert.Equal(t, Alphabet, FromNumerals([]uint16{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
		16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	}))
}
