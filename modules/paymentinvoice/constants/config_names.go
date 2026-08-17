package constants

import (
	core "github.com/sky-as-code/nikki-erp/modules/core/constants"
)

// Configuration keys of the Payment & Invoice module.
//
// Each gateway is gated by its own ENABLED flag, so a deployment that uses only one of them needs
// only that one's credentials. A gateway whose flag is false is not registered at all, and asking
// to pay through it fails as a client error rather than at the network.
//
// The three SECRET_KEY values and the two passwords are credentials: in a deployed environment
// they are supplied through the {KEY}_FILE secret-file convention, never written into
// config.default.yaml, which is committed and embedded into the binary.
const (
	MomoEnabled     core.ConfigName = "PAYMENTINVOICE.MOMO.ENABLED"
	MomoPartnerCode core.ConfigName = "PAYMENTINVOICE.MOMO.PARTNER_CODE"
	MomoAccessKey   core.ConfigName = "PAYMENTINVOICE.MOMO.ACCESS_KEY"
	MomoSecretKey   core.ConfigName = "PAYMENTINVOICE.MOMO.SECRET_KEY"
	MomoApiEndpoint core.ConfigName = "PAYMENTINVOICE.MOMO.API_ENDPOINT"

	// MomoIpnUrl is where MoMo posts the payment result. It is sent to MoMo on every create, so
	// changing it needs no action on MoMo's side, unlike the other two gateways' callbacks.
	MomoIpnUrl core.ConfigName = "PAYMENTINVOICE.MOMO.IPN_URL"

	// MomoRedirectUrl is where MoMo sends the payer's browser once they are done. It is a
	// human destination, not a callback: nothing is settled by a visit to it.
	MomoRedirectUrl core.ConfigName = "PAYMENTINVOICE.MOMO.REDIRECT_URL"
)

const (
	MposEnabled     core.ConfigName = "PAYMENTINVOICE.MPOS.ENABLED"
	MposMerchantId  core.ConfigName = "PAYMENTINVOICE.MPOS.MERCHANT_ID"
	MposSecretKey   core.ConfigName = "PAYMENTINVOICE.MPOS.SECRET_KEY"
	MposApiEndpoint core.ConfigName = "PAYMENTINVOICE.MPOS.API_ENDPOINT"
)

const (
	VietQrEnabled     core.ConfigName = "PAYMENTINVOICE.VIETQR.ENABLED"
	VietQrApiEndpoint core.ConfigName = "PAYMENTINVOICE.VIETQR.API_ENDPOINT"

	// The credentials this module presents when it calls VietQR.
	VietQrUsername  core.ConfigName = "PAYMENTINVOICE.VIETQR.USERNAME"
	VietQrPassword  core.ConfigName = "PAYMENTINVOICE.VIETQR.PASSWORD"
	VietQrSecretKey core.ConfigName = "PAYMENTINVOICE.VIETQR.SECRET_KEY"

	// The credentials the bank presents when it calls this module. VietQR's integration has the
	// partner host the token endpoint, so these are a separate pair from the two above and must
	// not be conflated with them.
	VietQrInboundUsername  core.ConfigName = "PAYMENTINVOICE.VIETQR.INBOUND_USERNAME"
	VietQrInboundPassword  core.ConfigName = "PAYMENTINVOICE.VIETQR.INBOUND_PASSWORD"
	VietQrInboundJwtSecret core.ConfigName = "PAYMENTINVOICE.VIETQR.INBOUND_JWT_SECRET"

	VietQrBankNumber core.ConfigName = "PAYMENTINVOICE.VIETQR.BANK_NUMBER"
	VietQrBankName   core.ConfigName = "PAYMENTINVOICE.VIETQR.BANK_NAME"
	VietQrBankCode   core.ConfigName = "PAYMENTINVOICE.VIETQR.BANK_CODE"
)

const (
	// OrderExpireAfterMins is how long an order may sit unpaid before the watchdog asks the
	// gateway for a verdict and, failing one, expires it.
	OrderExpireAfterMins core.ConfigName = "PAYMENTINVOICE.ORDER.EXPIRE_AFTER_MINS"

	// OrderCleanAfterHours is how long an unpaid or expired order is kept before deletion.
	OrderCleanAfterHours core.ConfigName = "PAYMENTINVOICE.ORDER.CLEAN_AFTER_HOURS"

	// SyncTimeoutSecs bounds one attempt to notify the ordering system of a payment result.
	SyncTimeoutSecs core.ConfigName = "PAYMENTINVOICE.SYNC.TIMEOUT_SECS"

	// SyncMaxRetries bounds how many times that notification is re-attempted.
	SyncMaxRetries core.ConfigName = "PAYMENTINVOICE.SYNC.MAX_RETRIES"
)
