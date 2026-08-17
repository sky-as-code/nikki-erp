package vietqr

import (
	"crypto/subtle"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.bryk.io/pkg/errors"
)

// InboundAuth is how the bank authenticates to *us*.
//
// VietQR's integration is unusual in having the partner host a token endpoint: the bank logs in
// with HTTP Basic, receives a bearer of our issuing, and presents it on every transaction
// callback. That makes this the mirror image of the adapter's own credentials, and the two pairs
// must never be conflated — presenting the bank's credentials back to it authenticates nobody.
type InboundAuth interface {
	// ValidateBasic checks the pair the bank presents at the token endpoint.
	ValidateBasic(username, password string) bool

	// IssueToken mints a bearer and reports its lifetime in seconds.
	IssueToken(username string) (token string, expiresIn int64, err error)

	// ValidateBearer checks the Authorization header of a callback.
	ValidateBearer(authorization string) bool
}

// tokenLifetime is how long an issued bearer stays valid.
//
// The bank re-authenticates on expiry, so this trades a login every hour against the window in
// which a leaked token is useful. It is deliberately not configurable: a deployment that set it
// to a year would have no way to revoke.
const tokenLifetime = time.Hour

// bearerPrefix is the scheme the Authorization header must carry, including its trailing space.
const bearerPrefix = "Bearer "

// InboundAuthImpl authenticates the bank against configured credentials.
//
// A deployment that has not configured the inbound pair gets an instance that refuses everything
// (see NewInboundAuth): an unset password must not become a blank one that any caller can match.
type InboundAuthImpl struct {
	username  string
	password  string
	jwtSecret []byte

	// configured is false when any of the three values above is missing. It is checked before
	// every credential comparison, so a half-configured deployment fails closed.
	configured bool
}

// NewInboundAuth builds the authenticator from the deployment's configured inbound credentials.
//
// All three values are required together. If any is absent the result refuses every request
// rather than falling back to a default: these credentials guard the endpoint that marks orders
// as paid, so a deployment that forgot to set them must reject the bank loudly, not accept
// anybody quietly.
func NewInboundAuth(username, password, jwtSecret string) *InboundAuthImpl {
	configured := username != "" && password != "" && jwtSecret != ""
	return &InboundAuthImpl{
		username:   username,
		password:   password,
		jwtSecret:  []byte(jwtSecret),
		configured: configured,
	}
}

// IsConfigured reports whether this deployment can serve the inbound endpoints at all. The wiring
// logs a warning when it cannot, so the cause is visible at boot rather than at the first callback.
func (this *InboundAuthImpl) IsConfigured() bool {
	return this.configured
}

// ValidateBasic compares the presented pair in constant time.
//
// subtle.ConstantTimeCompare rather than ==: string equality returns at the first differing byte,
// which leaks the length of the matching prefix and lets a caller recover the password one
// character at a time. Both fields are compared unconditionally, so the answer does not depend on
// which of the two was wrong.
func (this *InboundAuthImpl) ValidateBasic(username, password string) bool {
	if !this.configured {
		return false
	}

	userOk := subtle.ConstantTimeCompare([]byte(username), []byte(this.username)) == 1
	passOk := subtle.ConstantTimeCompare([]byte(password), []byte(this.password)) == 1
	return userOk && passOk
}

// IssueToken mints an HS256 bearer naming the caller.
func (this *InboundAuthImpl) IssueToken(username string) (string, int64, error) {
	if !this.configured {
		return "", 0, errors.New("vietqr inbound credentials are not configured")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
	})

	signed, err := token.SignedString(this.jwtSecret)
	if err != nil {
		return "", 0, errors.Wrap(err, "signing the vietqr inbound token")
	}

	return signed, int64(tokenLifetime.Seconds()), nil
}

// ValidateBearer checks that the header carries a bearer we issued and that has not expired.
//
// The signing method is pinned to HMAC. Without that check the library honours whatever `alg` the
// token itself declares, so a caller could present one signed with `none`, or an RS256 token whose
// "public key" is our HMAC secret, and have it verify. That is the standard JWT algorithm-
// confusion bypass, and here it would let anyone mark any order as paid.
func (this *InboundAuthImpl) ValidateBearer(authorization string) bool {
	if !this.configured {
		return false
	}

	tokenText, ok := strings.CutPrefix(authorization, bearerPrefix)
	if !ok || tokenText == "" {
		return false
	}

	token, err := jwt.ParseWithClaims(
		tokenText,
		&jwt.RegisteredClaims{},
		func(token *jwt.Token) (any, error) {
			if _, isHmac := token.Method.(*jwt.SigningMethodHMAC); !isHmac {
				return nil, errors.Errorf("unexpected signing method %q", token.Method.Alg())
			}
			return this.jwtSecret, nil
		},
		// Expiry is required rather than merely honoured when present: a token minted without one
		// would otherwise be valid forever.
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	return err == nil && token.Valid
}
