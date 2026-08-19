package vietqr

import (
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// The credential names a VietQR payment profile is written under. They are the spellings the
// service this module supersedes used, with snake_case aliases accepted because everything else on
// this side of the system is snake_case and an operator will reasonably try it.
const (
	profileKeyUsername   = "username"
	profileKeyPassword   = "password"
	profileKeySecretKey  = "secretKey"
	profileKeyBankCode   = "bankCode"
	profileKeyBankNumber = "bankNumber"
	profileKeyBankName   = "bankName"
)

// resolveConfig overlays a payment profile's credentials onto the deployment's.
//
// A VietQR profile is the one that most often carries everything: the bank account a transfer
// lands in is the account itself, so a second profile is usually a second account with its own
// login. Only what it supplies is overridden all the same, so a profile that names a bank account
// and nothing else still authenticates as this deployment does.
//
// The inbound credentials — the pair the bank presents when it calls us — are deliberately not
// here. They belong to the webhook layer and are per-deployment, not per-account; overriding them
// from a row would have this deployment expect a different bearer depending on which profile a
// transfer happened to be for, and the bank presents one bearer for all of them.
func (this *Adapter) resolveConfig(profileConfig map[string]any) Config {
	config := this.config
	if len(profileConfig) == 0 {
		return config
	}

	overrideIfSet(&config.Username, itGateway.ProfileString(profileConfig, profileKeyUsername, "user_name"))
	overrideIfSet(&config.Password, itGateway.ProfileString(profileConfig, profileKeyPassword))
	overrideIfSet(&config.SecretKey, itGateway.ProfileString(profileConfig, profileKeySecretKey, "secret_key"))
	overrideIfSet(&config.BankCode, itGateway.ProfileString(profileConfig, profileKeyBankCode, "bank_code"))
	overrideIfSet(&config.BankNumber, itGateway.ProfileString(profileConfig, profileKeyBankNumber, "bank_number"))
	overrideIfSet(&config.BankName, itGateway.ProfileString(profileConfig, profileKeyBankName, "bank_name"))

	return config
}

// overrideIfSet writes value over target unless the profile supplied nothing for it.
//
// The distinction matters: a profile that omits a credential means "use the deployment's", while
// writing an empty string over it would mean "this account has none", and the gateway would refuse
// the login or mint a QR code against no bank account at all.
func overrideIfSet(target *string, value string) {
	if value != "" {
		*target = value
	}
}
