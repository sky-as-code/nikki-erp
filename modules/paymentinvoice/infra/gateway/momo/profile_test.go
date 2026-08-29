package momo

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A payment profile is how one deployment collects into more than one MoMo merchant account. The
// rule these pin is that a profile overrides only what it supplies: anything it leaves out stays
// as configured, so a profile naming a partner code and nothing else still signs with this
// deployment's secret rather than with a blank one.

func TestAProfileOverridesOnlyWhatItSupplies(t *testing.T) {
	adapter := newTestAdapter()

	config := adapter.resolveConfig(map[string]any{
		"partnerCode": "MOMOSTORE02",
	})

	assert.Equal(t, "MOMOSTORE02", config.PartnerCode)
	assert.Equal(t, testAccessKey, config.AccessKey, "an omitted credential stays as configured")
	assert.Equal(t, testSecretKey, config.SecretKey, "an omitted credential stays as configured")
}

func TestAProfileMayCarryAWholeMerchantAccount(t *testing.T) {
	adapter := newTestAdapter()

	config := adapter.resolveConfig(map[string]any{
		"partnerCode": "MOMOSTORE02",
		"accessKey":   "other-access",
		"secretKey":   "other-secret",
		"storeId":     "HN-01",
		"redirectUrl": "https://example.test/done",
	})

	assert.Equal(t, "MOMOSTORE02", config.PartnerCode)
	assert.Equal(t, "other-access", config.AccessKey)
	assert.Equal(t, "other-secret", config.SecretKey)
	assert.Equal(t, "HN-01", config.StoreId)
	assert.Equal(t, "https://example.test/done", config.RedirectUrl)
}

// The service this module supersedes called the IPN destination returnUrl, so a profile migrated
// from it carries that name. Reading only "ipnUrl" would silently leave such a profile posting its
// results to whatever the deployment configured.
func TestTheOldReturnUrlSpellingIsAccepted(t *testing.T) {
	config := newTestAdapter().resolveConfig(map[string]any{
		"returnUrl": "https://example.test/ipn",
	})

	assert.Equal(t, "https://example.test/ipn", config.IpnUrl)
}

// An empty value is not an override. Writing it over the configured secret would have every
// signature computed against a blank key, and MoMo would refuse every request with nothing
// pointing at the row that caused it.
func TestAnEmptyProfileValueIsNotAnOverride(t *testing.T) {
	config := newTestAdapter().resolveConfig(map[string]any{
		"partnerCode": "",
		"secretKey":   "",
	})

	assert.Equal(t, testPartner, config.PartnerCode)
	assert.Equal(t, testSecretKey, config.SecretKey)
}

func TestNoProfileLeavesTheDeploymentConfigurationAlone(t *testing.T) {
	adapter := newTestAdapter()

	assert.Equal(t, adapter.config, adapter.resolveConfig(nil))
	assert.Equal(t, adapter.config, adapter.resolveConfig(map[string]any{}))
}

// MoMo signs its IPN with the secret of the account that took the money. Verifying a callback for
// a profile-collected payment against the deployment's own secret would reject a genuine payment
// as a forgery, which presents as money taken and goods never released.
func TestAnIpnIsVerifiedAgainstTheAccountThatTookTheMoney(t *testing.T) {
	adapter := newTestAdapter()
	profile := map[string]any{"accessKey": "other-access", "secretKey": "other-secret"}

	// The signature is recomputed from the fixture rather than written out, so this test pins who
	// signed the callback and not the fixture's field values, which signature_test.go already owns.
	payload := genuineIpnPayload()
	payload.Signature = signingFields{
		"accessKey":    "other-access",
		"amount":       strconv.FormatInt(payload.Amount, 10),
		"extraData":    payload.ExtraData,
		"message":      payload.Message,
		"orderId":      payload.OrderId,
		"orderInfo":    payload.OrderInfo,
		"orderType":    payload.OrderType,
		"partnerCode":  payload.PartnerCode,
		"payType":      payload.PayType,
		"requestId":    payload.RequestId,
		"responseTime": strconv.FormatInt(payload.ResponseTime, 10),
		"resultCode":   strconv.Itoa(payload.ResultCode),
		"transId":      strconv.FormatInt(payload.TransId, 10),
	}.sign("other-secret")

	assert.True(t, adapter.VerifyIpn(payload, profile),
		"a callback signed by the profile's account must verify")
	assert.False(t, adapter.VerifyIpn(payload, nil),
		"and must not verify against the deployment's own secret")
}
