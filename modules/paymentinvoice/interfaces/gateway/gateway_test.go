package gateway

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A payment profile's config is written by an operator, not by this codebase, and the same
// credential is spelled differently depending on which console wrote it. Reading only one spelling
// does not fail loudly: the credential simply looks absent, the deployment's own is used instead,
// and the money lands in the wrong merchant account.

func TestTheFirstKeyPresentWins(t *testing.T) {
	config := map[string]any{"partner_code": "SNAKE", "partnerCode": "CAMEL"}

	assert.Equal(t, "CAMEL", ProfileString(config, "partnerCode", "partner_code"))
	assert.Equal(t, "SNAKE", ProfileString(config, "partner_code", "partnerCode"))
}

func TestAnAliasIsReadWhenThePreferredSpellingIsAbsent(t *testing.T) {
	config := map[string]any{"partner_code": "SNAKE"}

	assert.Equal(t, "SNAKE", ProfileString(config, "partnerCode", "partner_code"))
}

// An empty value reads as absent, so the caller falls back to its configured credential rather
// than overriding it with nothing — a blank secret fails every signature the gateway checks.
func TestAnEmptyValueReadsAsAbsent(t *testing.T) {
	assert.Equal(t, "", ProfileString(map[string]any{"secretKey": ""}, "secretKey"))
	assert.Equal(t, "", ProfileString(map[string]any{"secretKey": nil}, "secretKey"))
}

// A value of the wrong type is not coerced. A number where a merchant id belongs is a config an
// operator has to fix, and quietly rendering it would hide that.
func TestANonStringValueIsNotCoerced(t *testing.T) {
	assert.Equal(t, "", ProfileString(map[string]any{"merchantId": 42}, "merchantId"))
}

func TestNoConfigReadsAsAbsent(t *testing.T) {
	assert.Equal(t, "", ProfileString(nil, "secretKey"))
	assert.Equal(t, "", ProfileString(map[string]any{}, "secretKey"))
}
