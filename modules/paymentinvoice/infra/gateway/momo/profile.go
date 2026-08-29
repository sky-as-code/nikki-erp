package momo

import (
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// The credential names a MoMo payment profile is written under.
//
// They are MoMo's own spellings, taken from its merchant portal, so an operator copying a value
// out of that portal files it under the name they read there. The snake_case aliases are accepted
// because everything else on this side of the system is snake_case and an operator will
// reasonably try it.
const (
	profileKeyPartnerCode = "partnerCode"
	profileKeyAccessKey   = "accessKey"
	profileKeySecretKey   = "secretKey"
	profileKeyStoreId     = "storeId"

	// profileKeyIpnUrl also accepts "returnUrl", which is what the service this module supersedes
	// called it. The two are the same thing — where MoMo posts the result — and a profile
	// migrated from that service carries the old name.
	profileKeyIpnUrl      = "ipnUrl"
	profileKeyReturnUrl   = "returnUrl"
	profileKeyRedirectUrl = "redirectUrl"
)

// resolveConfig overlays a payment profile's credentials onto the deployment's.
//
// Only what the profile actually supplies is overridden. A profile that names a partner code and
// nothing else is collected through that partner code under this deployment's secret, which is
// what a merchant with one contract and several store fronts has; a profile that names its own
// secret is a genuinely separate merchant account. An order that names no profile at all is
// collected exactly as it was before profiles existed.
//
// ApiEndpoint is deliberately not overridable. It is the host the adapter's HTTP caller was built
// around, so a profile could not change it here even if it named one, and letting a row redirect
// where credentials are sent would be a far larger power than holding them.
func (this *Adapter) resolveConfig(profile map[string]any) Config {
	config := this.config
	if len(profile) == 0 {
		return config
	}

	overrideIfSet(&config.PartnerCode, itGateway.ProfileString(profile, profileKeyPartnerCode, "partner_code"))
	overrideIfSet(&config.AccessKey, itGateway.ProfileString(profile, profileKeyAccessKey, "access_key"))
	overrideIfSet(&config.SecretKey, itGateway.ProfileString(profile, profileKeySecretKey, "secret_key"))
	overrideIfSet(&config.StoreId, itGateway.ProfileString(profile, profileKeyStoreId, "store_id"))
	overrideIfSet(&config.IpnUrl, itGateway.ProfileString(
		profile, profileKeyIpnUrl, "ipn_url", profileKeyReturnUrl, "return_url"))
	overrideIfSet(&config.RedirectUrl, itGateway.ProfileString(profile, profileKeyRedirectUrl, "redirect_url"))

	return config
}

// overrideIfSet writes value over target unless the profile supplied nothing for it.
//
// The distinction matters: a profile that omits a credential means "use the deployment's", while
// writing an empty string over it would mean "this account has none", and MoMo would refuse every
// request with a signature computed over a blank secret.
func overrideIfSet(target *string, value string) {
	if value != "" {
		*target = value
	}
}
