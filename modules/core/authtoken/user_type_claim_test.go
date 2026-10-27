package authtoken

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The claim must survive signing and parsing, which is the path a real token takes between the
// login service and the authorization middleware.
func roundTrip(t *testing.T, claims jwt.MapClaims) jwt.MapClaims {
	key := []byte("user-type-claim-test-key-0123456789")
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	require.NoError(t, err)

	parsed := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(signed, parsed, func(*jwt.Token) (any, error) { return key, nil })
	require.NoError(t, err)
	return parsed
}

func TestPrincipalKindFromClaims_RoundTrip(t *testing.T) {
	service := roundTrip(t, jwt.MapClaims(UserTypeClaims("service")))
	assert.Equal(t, "service", service[JwtClaimUserType])
	assert.Equal(t, corectx.PrincipalKindService, PrincipalKindFromClaims(service))

	user := roundTrip(t, jwt.MapClaims(UserTypeClaims("user")))
	assert.Equal(t, corectx.PrincipalKindUser, PrincipalKindFromClaims(user))
}

func TestPrincipalKindFromClaims_DefaultsToUser(t *testing.T) {
	assert.Equal(t, corectx.PrincipalKindUser, PrincipalKindFromClaims(jwt.MapClaims{"sub": "a@b.c:nikkiuser"}),
		"a token issued before the claim existed belongs to a person")
	assert.Equal(t, corectx.PrincipalKindUser, PrincipalKindFromClaims(jwt.MapClaims{JwtClaimUserType: 42}))
	assert.Equal(t, corectx.PrincipalKindUser, PrincipalKindFromClaims(&jwt.RegisteredClaims{}))
}
