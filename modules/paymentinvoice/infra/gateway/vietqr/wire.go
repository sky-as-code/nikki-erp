package vietqr

// The VietQR wire contract. Field names and JSON tags are the gateway's and the bank's, and may
// not be renamed — several of them are lower-case run-together words (`referencenumber`,
// `reftransactionid`) that no style guide would produce. That is the point: they are what the
// bank sends and expects.

// Endpoint paths. Note this deployment both calls VietQR and is called by the bank, so
// pathTokenGenerate appears on both sides of the integration: we POST to VietQR's copy to get a
// bearer, and we host our own copy for the bank to do the same.
const (
	pathTokenGenerate = "/vqr/api/token_generate"
	pathGenerateQr    = "/vqr/api/qr/generate-customer"
	pathRefund        = "/vqr/api/transaction/refund"
	pathCheckOrder    = "/vqr/api/transactions/check-order"
)

// QR types. A dynamic code carries the amount and is single-use, which is what a payment needs.
const (
	QrTypeDynamic     = 0
	QrTypeStatic      = 1
	QrTypeSemiDynamic = 3
)

// transTypeCredit marks the transaction as money coming in. The gateway spells debit "D".
const transTypeCredit = "C"

// Refund outcomes, as the gateway words them.
const (
	RefundStatusSuccess = "SUCCESS"
	RefundStatusFailed  = "FAILED"
)

// Check-order statuses.
const (
	CheckOrderStatusPending = 0
	CheckOrderStatusPaid    = 1
	CheckOrderStatusExpired = 2
)

// Error reasons in a webhook reply. The bank reads these two values and nothing else; "000" is
// success and "001" is "no such transaction".
const (
	WebhookReasonSuccess  = "000"
	WebhookReasonNotFound = "001"
)

// TokenResponse is the shape of both VietQR's token reply and the one we serve to the bank. The
// snake_case field names are OAuth-style and are what the bank parses.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// TokenTypeBearer is the only token type either side of this integration issues.
const TokenTypeBearer = "Bearer"

type generateQrRequest struct {
	BankCode     string `json:"bankCode"`
	BankAccount  string `json:"bankAccount"`
	UserBankName string `json:"userBankName"`

	// Content is capped at 23 characters by the gateway; longer text is rejected rather than
	// truncated, so the adapter trims it before sending.
	Content string `json:"content"`

	QrType  int    `json:"qrType"`
	Amount  int64  `json:"amount"`
	OrderId string `json:"orderId"`

	TransType    string `json:"transType"`
	TerminalCode string `json:"terminalcode"`
	Sign         string `json:"sign"`
	UrlLink      string `json:"urlLink"`
}

type generateQrResponse struct {
	BankCode         string `json:"bankCode"`
	BankName         string `json:"bankName"`
	BankAccount      string `json:"bankAccount"`
	UserBankName     string `json:"userBankName"`
	Amount           string `json:"amount"`
	Content          string `json:"content"`
	QrCode           string `json:"qrCode"`
	ImgId            string `json:"imgId"`
	TransactionId    string `json:"transactionId"`
	TransactionRefId string `json:"transactionRefId"`
	QrLink           string `json:"qrLink"`
	TerminalCode     string `json:"terminalCode"`
	OrderId          string `json:"orderId"`
	VaAccount        string `json:"vaAccount"`
}

type refundRequest struct {
	BankAccount     string `json:"bankAccount"`
	ReferenceNumber string `json:"referenceNumber"`

	// Amount is a string here, though the QR request sends it as a number. It is also the exact
	// text the checksum is computed over, so the two must be rendered identically.
	Amount string `json:"amount"`

	Content  string `json:"content"`
	CheckSum string `json:"checkSum"`
	BankCode string `json:"bankCode"`
}

type refundResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// checkOrderTypeByOrderId asks the gateway to look the transaction up by our order id rather than
// by its own reference number.
const checkOrderTypeByOrderId = 0

type checkOrderRequest struct {
	BankAccount string `json:"bankAccount"`
	Type        int    `json:"type"`
	Value       string `json:"value"`
	CheckSum    string `json:"checkSum"`
}

// checkOrderEntry is one transaction in the check-order reply, which is a JSON array.
type checkOrderEntry struct {
	ReferenceNumber string `json:"referenceNumber"`
	OrderId         string `json:"orderId"`
	Amount          int64  `json:"amount"`
	Content         string `json:"content"`
	TransType       string `json:"transType"`
	Status          int    `json:"status"`
	TimeCreated     int64  `json:"timeCreated"`
	TimePaid        int64  `json:"timePaid"`
	TerminalCode    string `json:"terminalCode"`
	Note            string `json:"note"`
	RefundCount     int    `json:"refundCount"`
	AmountRefunded  int64  `json:"amountRefunded"`
}

// WebhookPayload is the body the bank posts when a transfer arrives. The run-together lower-case
// names are the bank's and must not be tidied.
type WebhookPayload struct {
	BankAccount     string `json:"bankaccount"`
	Amount          int64  `json:"amount"`
	TransType       string `json:"transType"`
	Content         string `json:"content"`
	TransactionId   string `json:"transactionid"`
	TransactionTime int64  `json:"transactiontime"`
	ReferenceNumber string `json:"referencenumber"`
	OrderId         string `json:"orderId"`
}

// WebhookResponse is what the bank expects back, and its shape is non-standard: the outcome is
// carried in the body rather than the status code, and the bank keys off `errorReason`.
//
// Every field here matters to the caller, including `toastMessage`, which the bank displays to
// the teller. Reshaping this — returning a bare 200, renaming a field, dropping `object` when
// there is nothing to put in it — breaks the integration silently, because the bank will simply
// treat the reply as a failure and keep retrying.
type WebhookResponse struct {
	Error        bool              `json:"error"`
	ErrorReason  string            `json:"errorReason"`
	ToastMessage string            `json:"toastMessage"`
	Object       map[string]string `json:"object"`
}

// NewWebhookAccepted is the reply for a transfer we matched to an order and recorded.
func NewWebhookAccepted(refTransactionId string) WebhookResponse {
	return WebhookResponse{
		Error:        false,
		ErrorReason:  WebhookReasonSuccess,
		ToastMessage: "",
		Object:       map[string]string{"reftransactionid": refTransactionId},
	}
}

// NewWebhookNotFound is the reply for a transfer naming an order we do not have, or one already
// settled. The Vietnamese message is what the bank shows its teller, and is reproduced verbatim
// from the service this module supersedes.
func NewWebhookNotFound() WebhookResponse {
	return WebhookResponse{
		Error:        true,
		ErrorReason:  WebhookReasonNotFound,
		ToastMessage: "Giao dịch không tồn tại",
		Object:       map[string]string{"reftransactionid": ""},
	}
}
