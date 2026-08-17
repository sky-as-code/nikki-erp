package mpos

// The mPOS wire contract. Field names and JSON tags are the gateway's and may not be renamed.

// Service names. Every request names the operation it wants inside the encrypted body, rather
// than in the URL: the two paths below carry all of them between them.
const (
	ServiceAddOrder             = "ADD_ORDER_INFOR"
	ServiceGetTransactionStatus = "GET_TRANSACTION_STATUS"
	ServiceRefundTransaction    = "REFUND_TRANSACTION"
	ServiceRemoveOrder          = "REMOVE_ORDER_INFOR"

	// ServiceUpdateTransaction is what the gateway sends us, not what we send it.
	ServiceUpdateTransaction = "SERVICE_UPDATE_TRANSACTION"
)

// The two endpoint paths. Order-scoped services go to one, transaction-scoped to the other, and
// sending a service to the wrong path fails at the gateway.
const (
	pathOrder       = "/order"
	pathTransaction = "/transaction"
)

// ResponseCodeSuccess is the only code that means the request was carried out. The others named
// here are the few worth recognising in a log; the rest are reported as received.
const (
	ResponseCodeSuccess           = 200
	ResponseCodeUnableToDecrypt   = 1008
	ResponseCodeTransNotFound     = 3100
	ResponseCodeRequestNotFound   = 61111
	ResponseCodeAlreadyProcessed  = 61112
	ResponseCodeAmountMismatch    = 61118
	ResponseCodeFullyRefunded     = 33001
	ResponseCodeInvalidRefundAmnt = 9504
)

// Transaction statuses reported by the terminal.
//
// Three of them mean the payment succeeded and the distinction is about settlement, not outcome:
// APPROVED is authorised, SETTLED is authorised and settled by the bank (and is what a QR payment
// reports), and PENDING_SIGNATURE is authorised while the terminal still wants the cardholder's
// signature. Treating only APPROVED as paid would strand real payments — see isPaid.
const (
	TransStatusPending          = 90
	TransStatusRejected         = 91
	TransStatusFailed           = 97
	TransStatusRefund           = 99
	TransStatusApproved         = 100
	TransStatusReversed         = 101
	TransStatusVoided           = 102
	TransStatusPendingSignature = 103
	TransStatusSettled          = 104
)

// Payment methods the terminal can be told to present.
const (
	// PaymentMethodCard makes the terminal show its card-swipe prompt as soon as the order
	// arrives. It is the default this adapter asks for.
	PaymentMethodCard = "CARD"
	PaymentMethodQr   = "QR"
	PaymentMethodLink = "LINK"
)

// paymentTypeOther is the non-installment payment type. The gateway spells the value exactly as
// below, spaces and all.
const paymentTypeOther = "Other values"

// envelope is the outer, unencrypted body of every request: the merchant id in the clear so the
// gateway knows which secret to decrypt with, and the real payload inside reqData.
type envelope struct {
	MerchantId string `json:"merchantId"`
	ReqData    string `json:"reqData"`
}

// WebhookEnvelope is the same shape arriving the other way, when the terminal calls us. It is
// exported because the webhook handler decodes it before handing ReqData here to be decrypted.
type WebhookEnvelope = envelope

// response is the outer reply. resData is encrypted the same way the request was.
type response struct {
	MerchantId string `json:"merchantId"`
	ResCode    int    `json:"resCode"`
	ResData    string `json:"resData"`
	Message    string `json:"message"`
}

type addOrderRequest struct {
	ServiceName   string `json:"serviceName"`
	OrderId       string `json:"orderId"`
	PosId         string `json:"posId"`
	Amount        int64  `json:"amount"`
	Description   string `json:"description,omitempty"`
	PaymentType   string `json:"paymentType"`
	PaymentMethod string `json:"paymentMethod"`
}

type removeOrderRequest struct {
	ServiceName string `json:"serviceName"`
	OrderId     string `json:"orderId"`
	PosId       string `json:"posId"`
	Amount      int64  `json:"amount"`
}

// checkOrderRequest sends the amount as a string, unlike every other request here. That is the
// gateway's contract, not an oversight.
type checkOrderRequest struct {
	ServiceName string `json:"serviceName"`
	OrderId     string `json:"orderId"`
	PosId       string `json:"posId"`
	Amount      string `json:"amount"`
}

type checkOrderResponse struct {
	ServiceName string `json:"serviceName"`
	OrderId     string `json:"orderId"`
	PosId       string `json:"posId"`
	Amount      string `json:"amount"`
	TransStatus int    `json:"transStatus"`
	IssuerCode  string `json:"issuerCode"`
	TransCode   string `json:"transCode"`
	TransDate   string `json:"transDate"`
}

type refundRequest struct {
	ServiceName  string `json:"serviceName"`
	TransCode    string `json:"transCode"`
	RequestId    string `json:"requestId"`
	RefundAmount int64  `json:"refundAmount"`
}

type refundResponse struct {
	ServiceName  string `json:"serviceName"`
	TransCode    string `json:"transCode"`
	RefundAmount int64  `json:"refundAmount"`
	RestedAmount int64  `json:"restedAmount"`
	RequestId    string `json:"requestId"`
}

// WebhookPayload is the decrypted body the terminal sends when a transaction changes state. It is
// exported because the webhook handler hands the raw reqData here to be decrypted into it.
type WebhookPayload struct {
	ServiceName string `json:"serviceName"`
	TransStatus int    `json:"transStatus"`
	TransCode   string `json:"transCode"`
	TransDate   string `json:"transDate"`
	TransAmount int64  `json:"transAmount"`
	IssuerCode  string `json:"issuerCode"`
	Muid        string `json:"muid"`
	OrderId     string `json:"orderId"`
	PosId       string `json:"posId"`
}

// isPaid reports whether a transaction status means the money was actually taken.
//
// All three of these are successful outcomes: the service this replaces only ever compared
// against APPROVED, which would leave a settled QR payment — status 104 — looking unpaid and
// expire an order the customer had already paid for.
func isPaid(transStatus int) bool {
	return transStatus == TransStatusApproved ||
		transStatus == TransStatusSettled ||
		transStatus == TransStatusPendingSignature
}

// isSettled reports whether the gateway has reached a verdict at all. A pending transaction is
// still in flight and must be left alone rather than treated as a failure.
func isSettled(transStatus int) bool {
	return transStatus != TransStatusPending
}
