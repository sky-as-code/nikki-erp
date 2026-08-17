package vietqr

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These credentials guard the endpoint that marks an order paid. Every test below is a way
// someone could otherwise get past it.

const (
	testUser   = "bank-user"
	testPass   = "bank-password"
	testSecret = "an-inbound-signing-secret"
)

func newTestAuth() *InboundAuthImpl {
	return NewInboundAuth(testUser, testPass, testSecret)
}

func TestTheConfiguredPairIsAccepted(t *testing.T) {
	assert.True(t, newTestAuth().ValidateBasic(testUser, testPass))
}

func TestAWrongCredentialIsRefused(t *testing.T) {
	auth := newTestAuth()

	for name, pair := range map[string][2]string{
		"wrong password": {testUser, "nope"},
		"wrong username": {"nobody", testPass},
		"both wrong":     {"nobody", "nope"},
		"both empty":     {"", ""},
		"empty password": {testUser, ""},
		// A prefix of the real password must not pass: this is what a comparison that stopped at
		// the first differing byte would leak the length of.
		"password prefix": {testUser, testPass[:5]},
	} {
		assert.False(t, auth.ValidateBasic(pair[0], pair[1]), name)
	}
}

// A deployment that never set the inbound credentials must refuse the bank outright. The trap is
// the empty string: if unset config were compared literally, a caller presenting "" for both
// fields would match and gain the authority to settle any order.
func TestAnUnconfiguredDeploymentRefusesEveryone(t *testing.T) {
	for name, auth := range map[string]*InboundAuthImpl{
		"nothing set":   NewInboundAuth("", "", ""),
		"no password":   NewInboundAuth(testUser, "", testSecret),
		"no jwt secret": NewInboundAuth(testUser, testPass, ""),
		"no username":   NewInboundAuth("", testPass, testSecret),
	} {
		assert.False(t, auth.IsConfigured(), name)
		assert.False(t, auth.ValidateBasic("", ""), name+": blank pair")
		assert.False(t, auth.ValidateBasic(testUser, testPass), name+": real pair")

		_, _, err := auth.IssueToken(testUser)
		assert.Error(t, err, name+": must not mint a token")
	}
}

func TestAnIssuedTokenValidates(t *testing.T) {
	auth := newTestAuth()

	token, expiresIn, err := auth.IssueToken(testUser)

	require.NoError(t, err)
	assert.Equal(t, int64(3600), expiresIn)
	assert.True(t, auth.ValidateBearer(bearerPrefix+token))
}

func TestAMalformedAuthorizationHeaderIsRefused(t *testing.T) {
	auth := newTestAuth()
	token, _, err := auth.IssueToken(testUser)
	require.NoError(t, err)

	for name, header := range map[string]string{
		"empty":            "",
		"no scheme":        token,
		"wrong scheme":     "Basic " + token,
		"lower-case bearer": "bearer " + token,
		"scheme only":      bearerPrefix,
		"not a token":      bearerPrefix + "garbage",
		"no space":         "Bearer" + token,
	} {
		assert.False(t, auth.ValidateBearer(header), name)
	}
}

// A token signed with a different secret must not verify — otherwise anyone who could mint one
// could settle orders.
func TestATokenSignedWithAnotherSecretIsRefused(t *testing.T) {
	issuer := NewInboundAuth(testUser, testPass, "some-other-secret")
	token, _, err := issuer.IssueToken(testUser)
	require.NoError(t, err)

	assert.False(t, newTestAuth().ValidateBearer(bearerPrefix+token))
}

// The algorithm-confusion bypass. A parser that honours whatever `alg` the token declares accepts
// one signed with "none" — no key, no signature, trivially forged by anyone who knows the URL.
func TestAnUnsignedTokenIsRefused(t *testing.T) {
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.RegisteredClaims{
		Subject:   testUser,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	assert.False(t, newTestAuth().ValidateBearer(bearerPrefix+unsigned))
}

// The other half of the same bypass: a stronger-looking HMAC variant is still not the one we
// issue, and accepting it widens what an attacker may present.
func TestATokenSignedWithAnotherHmacVariantIsRefused(t *testing.T) {
	other, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Subject:   testUser,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	assert.False(t, newTestAuth().ValidateBearer(bearerPrefix+other))
}

func TestAnExpiredTokenIsRefused(t *testing.T) {
	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   testUser,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	assert.False(t, newTestAuth().ValidateBearer(bearerPrefix+expired))
}

// A token with no expiry at all would otherwise be valid forever, which is indistinguishable from
// having no revocation.
func TestATokenWithoutAnExpiryIsRefused(t *testing.T) {
	eternal, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: testUser,
	}).SignedString([]byte(testSecret))
	require.NoError(t, err)

	assert.False(t, newTestAuth().ValidateBearer(bearerPrefix+eternal))
}
