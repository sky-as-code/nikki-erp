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

	// MomoStoreId names the merchant's store, for accounts whose settlement MoMo breaks down by
	// one. Optional: an account with no such breakdown leaves it unset and it is not sent.
	MomoStoreId core.ConfigName = "PAYMENTINVOICE.MOMO.STORE_ID"
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

// PaymentSettledEventTopic is where this module announces that an order reached a verdict.
//
// Distinct from the SYNC keys above, which configure the HTTP callback to an ordering system that
// asked for one. This is the in-process announcement any module in the same build can subscribe to,
// and the two are independent: an order can have both, either, or neither.
const PaymentSettledEventTopic core.ConfigName = "PAYMENTINVOICE.EVENT.PAYMENT_SETTLED_TOPIC"

// DefaultPaymentSettledEventTopic is the fallback when PaymentSettledEventTopic is unset.
//
// Declared once here and read by both the publisher and every subscriber, so the two halves of a
// pub/sub pair cannot drift apart over where they meet — the mistake the vending-machine module
// made by declaring its topic twice.
const DefaultPaymentSettledEventTopic = "nikkierp.paymentinvoice.events.payment_settled"

// PaymentProfileEncryptionKey names the hex-encoded 32-byte AES key that a payment profile's
// gateway credentials are encrypted with before they reach the database.
//
// It is deliberately the core key rather than one of this module's own. Coremart's vending-machine
// module encrypts its payment configs with the same key, so a profile can be moved between the two
// without being re-encrypted, and neither side can be rotated into a state where it cannot read
// what the other wrote.
//
// It is a credential: supply it through config/local.env or the {KEY}_FILE secret-file convention,
// never in config.default.yaml, which is committed and embedded into the binary.
const PaymentProfileEncryptionKey core.ConfigName = "CORE.ENCRYPTION.EAS_32_BYTES_KEY"
