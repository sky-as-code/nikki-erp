package mpos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mPOS is the gateway where the credentials matter twice over: the merchant secret is both what
// authenticates a request and what encrypts it, so a profile resolved to the wrong account does
// not fail cleanly — it produces a body the gateway cannot read at all.

func TestAProfileOverridesOnlyWhatItSupplies(t *testing.T) {
	config, err := newTestAdapter().resolveConfig(map[string]any{"merchantId": "M2"})

	require.NoError(t, err)
	assert.Equal(t, "M2", config.MerchantId)
	assert.Equal(t, testSecretKey, config.SecretKey, "an omitted credential stays as configured")
}

func TestAProfileMayCarryAWholeMerchantAccount(t *testing.T) {
	config, err := newTestAdapter().resolveConfig(map[string]any{
		"merchantId": "M2",
		"secretKey":  "0123456789abcdef",
	})

	require.NoError(t, err)
	assert.Equal(t, "M2", config.MerchantId)
	assert.Equal(t, "0123456789abcdef", config.SecretKey)
}

// The secret is used directly as the AES-128 key, so a wrong length is not a credential that
// merely fails to authenticate. Refusing it here names the profile as the problem; letting it
// through would surface as a body that "cannot be decrypted", which is indistinguishable from a
// forged callback.
func TestAProfileSecretOfTheWrongLengthIsRefused(t *testing.T) {
	_, err := newTestAdapter().resolveConfig(map[string]any{"secretKey": "too-short"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "16 characters")
}

func TestNoProfileLeavesTheDeploymentConfigurationAlone(t *testing.T) {
	adapter := newTestAdapter()

	config, err := adapter.resolveConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, adapter.config, config)
}

// An inbound callback carries the merchant id in the clear and everything else encrypted under
// that account's secret, so the account has to be found by this id before the body can be read.
func TestMerchantIdIsReadableFromAProfile(t *testing.T) {
	assert.Equal(t, "M2", MerchantIdOf(map[string]any{"merchantId": "M2"}))
	assert.Equal(t, "M2", MerchantIdOf(map[string]any{"merchant_id": "M2"}))
	assert.Equal(t, "", MerchantIdOf(nil))
	assert.Equal(t, "", MerchantIdOf(map[string]any{"secretKey": "0123456789abcdef"}))
}

// A callback encrypted by one account must not be readable with another's secret, and the failure
// has to be an error rather than nonsense that decodes.
func TestACallbackIsReadOnlyByTheAccountThatSentIt(t *testing.T) {
	adapter := newTestAdapter()
	profile := map[string]any{"merchantId": "M2", "secretKey": "0123456789abcdef"}

	reqData, err := encrypt(WebhookPayload{OrderId: "ORD1234ABCD5", TransCode: "TX1"}, "0123456789abcdef")
	require.NoError(t, err)

	payload, err := adapter.DecryptWebhook(reqData, profile)
	require.NoError(t, err)
	assert.Equal(t, "ORD1234ABCD5", payload.OrderId)

	_, err = adapter.DecryptWebhook(reqData, nil)
	assert.Error(t, err, "the deployment's own secret must not read another account's callback")
}
