package vietqr

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// VietQR is the gateway where a second merchant account is usually a second bank account, so a
// profile here typically carries everything: its own login, its own secret and its own account
// number. Two things have to hold for that to be safe — a profile overrides only what it names,
// and a bearer issued to one login is never presented on another's request.

func TestAProfileOverridesOnlyWhatItSupplies(t *testing.T) {
	config := newTestAdapter().resolveConfig(map[string]any{"bankNumber": "9999999999"})

	assert.Equal(t, "9999999999", config.BankNumber)
	assert.Equal(t, testUsername, config.Username, "an omitted credential stays as configured")
	assert.Equal(t, testSecretKey, config.SecretKey, "an omitted credential stays as configured")
}

func TestAProfileMayCarryAWholeBankAccount(t *testing.T) {
	config := newTestAdapter().resolveConfig(map[string]any{
		"username":   "other-user",
		"password":   "other-password",
		"secretKey":  "other-secret",
		"bankCode":   "VCB",
		"bankNumber": "9999999999",
		"bankName":   "OTHER MERCHANT",
	})

	assert.Equal(t, "other-user", config.Username)
	assert.Equal(t, "other-password", config.Password)
	assert.Equal(t, "other-secret", config.SecretKey)
	assert.Equal(t, "VCB", config.BankCode)
	assert.Equal(t, "9999999999", config.BankNumber)
	assert.Equal(t, "OTHER MERCHANT", config.BankName)
}

func TestNoProfileLeavesTheDeploymentConfigurationAlone(t *testing.T) {
	adapter := newTestAdapter()

	assert.Equal(t, adapter.config, adapter.resolveConfig(nil))
	assert.Equal(t, adapter.config, adapter.resolveConfig(map[string]any{}))
}

// One cache shared across accounts would hand a profile's request the bearer of whichever account
// authenticated first, and VietQR would then mint the QR code against that other merchant's bank
// account — the customer pays, and the money lands in the wrong place.
func TestABearerIsNeverReusedAcrossAccounts(t *testing.T) {
	store := newTokenStore(time.Now)
	logins := 0
	login := func() (string, time.Duration, error) {
		logins++
		return "token-for-login-" + string(rune('0'+logins)), time.Hour, nil
	}

	first, err := store.cacheFor("merchant-a").get(login)
	require.NoError(t, err)
	second, err := store.cacheFor("merchant-b").get(login)
	require.NoError(t, err)

	assert.NotEqual(t, first, second, "each account authenticates for itself")
	assert.Equal(t, 2, logins)

	// The same account is still served from cache: the point is separation, not re-authenticating
	// on every call, which is the behaviour this cache exists to remove.
	again, err := store.cacheFor("merchant-a").get(login)
	require.NoError(t, err)
	assert.Equal(t, first, again)
	assert.Equal(t, 2, logins)
}
