package codify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodify(t *testing.T) {
	cases := map[string]string{
		"Nước ngọt có ga":       "nuoc_ngot_co_ga",
		"Đồ uống / Bia & Rượu":  "do_uong_bia_ruou",
		"  Snack   Foods  ":     "snack_foods",
		"Café-Crème (Premium)":  "cafe_creme_premium",
		"ABC123":                "abc123",
		"":                      "",
		"---":                   "",
		"Điện thoại thông minh": "dien_thoai_thong_minh",
	}
	for input, want := range cases {
		assert.Equal(t, want, Codify(input), input)
	}
}

func TestCodifyTruncatesOnAWordBoundary(t *testing.T) {
	code := Codify("alpha beta gamma delta epsilon", 17)

	assert.Equal(t, "alpha_beta_gamma", code)
	assert.LessOrEqual(t, len(Codify("abcdefghijklmnopqrstuvwxyz", 10)), 10)
}
