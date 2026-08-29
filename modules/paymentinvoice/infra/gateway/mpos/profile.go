package mpos

import (
	"go.bryk.io/pkg/errors"

	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// The credential names an mPOS payment profile is written under. They are NextPay's own spellings,
// with snake_case aliases accepted because everything else on this side of the system is
// snake_case and an operator will reasonably try it.
const (
	profileKeyMerchantId = "merchantId"
	profileKeySecretKey  = "secretKey"
)

// MerchantIdOf reports which merchant account a payment profile names, or "" when it names none.
//
// It is exported because an inbound callback carries the merchant id in the clear and nothing
// else: the order it concerns is inside the encrypted body, so the profile has to be found by this
// id before anything in the callback can be read at all.
func MerchantIdOf(profileConfig map[string]any) string {
	return itGateway.ProfileString(profileConfig, profileKeyMerchantId, "merchant_id")
}

// resolveConfig overlays a payment profile's credentials onto the deployment's.
//
// Only what the profile supplies is overridden, so a profile carrying a merchant id and nothing
// else still uses this deployment's secret. An order that names no profile is served exactly as it
// was before profiles existed.
//
// A secret of the wrong length is refused here rather than passed on. It is used directly as the
// AES-128 key, so the failure would otherwise surface from inside the cipher — as an opaque error
// on a payment, or, on the callback path, as a body that merely "will not decrypt", which is
// indistinguishable from a forged one.
func (this *Adapter) resolveConfig(profileConfig map[string]any) (Config, error) {
	config := this.config
	if len(profileConfig) == 0 {
		return config, nil
	}

	if merchantId := MerchantIdOf(profileConfig); merchantId != "" {
		config.MerchantId = merchantId
	}
	if secretKey := itGateway.ProfileString(profileConfig, profileKeySecretKey, "secret_key"); secretKey != "" {
		if len(secretKey) != SecretKeyLength {
			return Config{}, errors.Errorf(
				"mpos payment profile secret key must be exactly %d characters, got %d",
				SecretKeyLength, len(secretKey))
		}
		config.SecretKey = secretKey
	}

	return config, nil
}
