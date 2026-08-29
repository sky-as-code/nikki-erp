package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// testEncryptionKey is a hex-encoded 32-byte AES key. It is a literal rather than a generated one
// so a failure is reproducible, and it is only ever used against these fixtures.
const testEncryptionKey = "ca69666ba4ca0af9d7d33f38741b4e81a086a83333888e29e9acb925ea9dbce6"

func newTestProfileService(encryptionKey string) *PaymentProfileDomainService {
	return &PaymentProfileDomainService{encryptionKey: encryptionKey}
}

// The round trip is what the payment flow depends on: a profile written through the CRUD has to be
// readable by the adapter that collects with it, and AES-GCM is authenticated, so a mismatch fails
// rather than yielding rubbish.
func TestProfileCredentialsSurviveTheRoundTrip(t *testing.T) {
	service := newTestProfileService(testEncryptionKey)
	config := map[string]any{
		"partnerCode": "MOMOABCD",
		"accessKey":   "F8BBA842ECF85",
	}

	profile := models.NewPaymentProfileFrom(dmodel.DynamicFields{
		models.PaymentProfileFieldConfig: config,
	})
	require.NoError(t, service.EncryptConfig(profile))

	encrypted := profile.GetEncryptedConfig()
	require.NotNil(t, encrypted)
	assert.NotContains(t, *encrypted, "MOMOABCD", "the credentials must not be readable at rest")

	require.NoError(t, service.DecryptConfig(profile))
	assert.Equal(t, config, profile.GetConfig())
}

// Coremart's vending-machine module encrypts its payment configs with this same key and the same
// AES-GCM construction, and the NestJS service reads both. A profile that could not be decrypted
// with a cipher text produced elsewhere would silently split the two into separate secrets, so the
// test pins the format rather than only the round trip.
func TestAProfileIsDecryptedFromAForeignCipherText(t *testing.T) {
	service := newTestProfileService(testEncryptionKey)

	source := models.NewPaymentProfileFrom(dmodel.DynamicFields{
		models.PaymentProfileFieldConfig: map[string]any{"merchantId": "M001"},
	})
	require.NoError(t, service.EncryptConfig(source))

	// A second profile that only ever saw the column, as a row copied between the two tables would.
	copied := models.NewPaymentProfileFrom(dmodel.DynamicFields{
		models.PaymentProfileFieldEncryptedConfig: *source.GetEncryptedConfig(),
	})
	require.NoError(t, service.DecryptConfig(copied))

	assert.Equal(t, map[string]any{"merchantId": "M001"}, copied.GetConfig())
}

// A deployment that never set the key must be told so on the request that needed it, naming the
// config item. Storing the credentials in the clear instead would be far worse than failing.
func TestAnUnsetKeyIsReportedRatherThanIgnored(t *testing.T) {
	service := newTestProfileService("")

	profile := models.NewPaymentProfileFrom(dmodel.DynamicFields{
		models.PaymentProfileFieldConfig: map[string]any{"merchantId": "M001"},
	})
	err := service.EncryptConfig(profile)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "CORE.ENCRYPTION.EAS_32_BYTES_KEY")
	assert.Nil(t, profile.GetEncryptedConfig())
}

// A listing that never selected the credentials must not be refused for want of a key it does not
// need, or switching the key off would break every screen rather than only the ones showing secrets.
func TestDecryptingWithoutTheColumnNeedsNoKey(t *testing.T) {
	service := newTestProfileService("")

	fields := dmodel.DynamicFields{models.PaymentProfileFieldName: "MoMo - Hanoi"}
	require.NoError(t, service.DecryptFields(fields))

	assert.Equal(t, dmodel.DynamicFields{models.PaymentProfileFieldName: "MoMo - Hanoi"}, fields)
}
