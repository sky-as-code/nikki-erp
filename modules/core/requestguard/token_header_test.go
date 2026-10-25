package requestguard

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestExtractPrefixedToken(t *testing.T) {
	cases := []struct {
		name   string
		header string
		prefix string
		want   string
	}{
		{"empty prefix takes the raw token", "abc.def.ghi", "", "abc.def.ghi"},
		{"empty prefix trims whitespace", "  abc.def.ghi  ", " ", "abc.def.ghi"},
		{"bearer prefix", "Bearer abc.def.ghi", "Bearer", "abc.def.ghi"},
		{"prefix is case-insensitive", "bearer abc.def.ghi", "Bearer", "abc.def.ghi"},
		{"prefix needs a separator", "Bearerabc.def.ghi", "Bearer", ""},
		{"missing prefix", "abc.def.ghi", "Bearer", ""},
		{"prefix only", "Bearer", "Bearer", ""},
		{"empty header", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, extractPrefixedToken(tc.header, tc.prefix))
		})
	}
}

func TestClaimScopes(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, claimScopes(jwt.MapClaims{"scope": "a  b"}))
	assert.Equal(t, []string{"a", "b"}, claimScopes(jwt.MapClaims{"scope": []any{"a", "b", 3}}))
	assert.Nil(t, claimScopes(jwt.MapClaims{}))
}

func TestScopesIntersect(t *testing.T) {
	assert.True(t, scopesIntersect([]string{"read", "write"}, []string{"write"}))
	assert.False(t, scopesIntersect([]string{"read"}, []string{"write"}))
	assert.False(t, scopesIntersect(nil, []string{"write"}))
	assert.False(t, scopesIntersect([]string{"read"}, nil))
}
