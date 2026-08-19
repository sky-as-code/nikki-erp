package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A payment profile's whole reason to exist is that the merchant credentials are readable in a
// request and unreadable in the database. These pin the two directions of that swap, and the cases
// where it must do nothing at all — a partial update that never mentioned the credentials, and a
// listing that never selected them.

// reverseEncrypt stands in for the real cipher. The tests care that the plain text is handed over,
// that the result is stored, and that the plain field is gone afterwards — none of which depends on
// the cipher being real, and a stub makes a wrong direction obvious in the assertion.
func reverseEncrypt(plain string) (string, error) {
	runes := []rune(plain)
	for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
		runes[left], runes[right] = runes[right], runes[left]
	}
	return string(runes), nil
}

func TestConfigIsNotASchemaField(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := PaymentProfileSchemaBuilder().Build()

	_, hasConfig := schema.Fields()[PaymentProfileFieldConfig]
	assert.False(t, hasConfig,
		"declaring 'config' would give it a column, which is the very thing the encryption avoids")

	_, hasEncrypted := schema.Fields()[PaymentProfileFieldEncryptedConfig]
	assert.True(t, hasEncrypted, "the encrypted form is the only one that is stored")
}

func TestParseConfigToEncryptedConfigReplacesThePlainField(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldConfig: map[string]any{"partnerCode": "MOMOABCD"},
	})

	require.NoError(t, profile.ParseConfigToEncryptedConfig(reverseEncrypt))

	fields := profile.GetFieldData()
	_, stillPlain := fields[PaymentProfileFieldConfig]
	assert.False(t, stillPlain, "the plain config must never reach the repository")

	encrypted := profile.GetEncryptedConfig()
	require.NotNil(t, encrypted)
	assert.Equal(t, `}"DCBAOMOM":"edoCrentrap"{`, *encrypted)
}

// A partial update that renames a profile carries no config key at all. Writing an absent field as
// null would wipe the credentials of every profile edited for its name alone.
func TestAnAbsentConfigIsLeftAlone(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldName: "MoMo - Hanoi",
	})

	require.NoError(t, profile.ParseConfigToEncryptedConfig(reverseEncrypt))

	_, wrote := profile.GetFieldData()[PaymentProfileFieldEncryptedConfig]
	assert.False(t, wrote, "a request that did not mention the credentials must not rewrite them")
}

// A config key present but null is the caller clearing the credentials, which is a different
// statement from not mentioning them and is honoured as one.
func TestAnExplicitNullConfigClearsTheCredentials(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldConfig: nil,
	})

	require.NoError(t, profile.ParseConfigToEncryptedConfig(reverseEncrypt))

	fields := profile.GetFieldData()
	value, wrote := fields[PaymentProfileFieldEncryptedConfig]
	assert.True(t, wrote)
	assert.Nil(t, value)
}

func TestParseEncryptedConfigToConfigDropsTheCipherText(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldEncryptedConfig: `}"DCBAOMOM":"edoCrentrap"{`,
	})

	require.NoError(t, profile.ParseEncryptedConfigToConfig(reverseEncrypt))

	fields := profile.GetFieldData()
	_, stillEncrypted := fields[PaymentProfileFieldEncryptedConfig]
	assert.False(t, stillEncrypted, "the cipher text must never be exposed to a caller")

	assert.Equal(t, map[string]any{"partnerCode": "MOMOABCD"}, profile.GetConfig())
}

// A listing that selected only name and method has no encrypted_config key. Decryption must be a
// no-op there rather than an error, or every listing would fail for want of a column it never
// asked for.
func TestDecryptingARecordWithoutTheColumnIsANoOp(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldName: "MoMo - Hanoi",
	})

	require.NoError(t, profile.ParseEncryptedConfigToConfig(func(string) (string, error) {
		t.Fatal("decryptFn must not be called when there is nothing to decrypt")
		return "", nil
	}))

	assert.Nil(t, profile.GetConfig())
}

// The method values are what an order's gateway is selected by, and the first three are the same
// strings as the adapter codes. A value here that no adapter answers to is a profile that can take
// a payment nobody can collect.
func TestProfileMethodsMatchTheAdapterCodes(t *testing.T) {
	assert.Equal(t, AdapterCodeMomo, string(PaymentProfileMethodMomo))
	assert.Equal(t, AdapterCodeVietQr, string(PaymentProfileMethodVietQr))
	assert.Equal(t, AdapterCodeMpos, string(PaymentProfileMethodMpos))
}

// The operator console writes a profile's config as a table of {key, value} rows, so what comes
// back out of the cipher is not the flat object an adapter wants to read. These pin the flattening
// in both directions, because a credential that silently fails to flatten does not error — the
// adapter simply falls back to the deployment's own, and the payment lands in the wrong account.
func TestConfigValuesFlattenTheConsolesKeyValueRows(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldConfig: map[string]any{
			"0": map[string]any{"key": "partnerCode", "value": "MOMOSTORE02"},
			"1": map[string]any{"key": "accessKey", "value": "F8BBA842ECF85"},
		},
	})

	assert.Equal(t, map[string]any{
		"partnerCode": "MOMOSTORE02",
		"accessKey":   "F8BBA842ECF85",
	}, profile.ConfigValues())
}

// The same rows arrive as a JSON array when the console writes them as a list rather than an
// object. Both shapes are in the field today, so both have to flatten to the same thing.
func TestConfigValuesFlattenAListOfRows(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldConfig: []any{
			map[string]any{"key": "merchantId", "value": "M2"},
			map[string]any{"key": "secretKey", "value": "0123456789abcdef"},
		},
	})

	assert.Equal(t, map[string]any{
		"merchantId": "M2",
		"secretKey":  "0123456789abcdef",
	}, profile.ConfigValues())
}

// A profile written by hand or by an import is a plain object already, and must be read as it
// stands rather than treated as malformed.
func TestConfigValuesPassAPlainObjectThrough(t *testing.T) {
	profile := NewPaymentProfileFrom(dmodel.DynamicFields{
		PaymentProfileFieldConfig: map[string]any{"bankNumber": "9999999999"},
	})

	assert.Equal(t, map[string]any{"bankNumber": "9999999999"}, profile.ConfigValues())
}

func TestConfigValuesOfAnUnreadProfileIsNil(t *testing.T) {
	assert.Nil(t, NewPaymentProfile().ConfigValues())
}
