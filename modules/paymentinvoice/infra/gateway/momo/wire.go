package momo

// The MoMo wire contract. Field names and JSON tags are MoMo's and may not be renamed.

// ResultCodeSuccess is the only result code that means the operation succeeded. Every other value
// is a failure of some kind; the ones named below are the few this adapter branches on, and the
// rest are reported as-is rather than being enumerated for their own sake.
const (
	ResultCodeSuccess         = 0
	ResultCodeDuplicateOrder  = 41
	ResultCodeTransNotFound   = 1005
	ResultCodeUserCancelled   = 1006
	ResultCodeTransPending    = 1000
	ResultCodeTransProcessing = 7000
)

// requestTypeCaptureWallet asks MoMo for the wallet flow, which is what returns both a QR code
// and a pay URL. It is the only flow this adapter uses.
const requestTypeCaptureWallet = "captureWallet"

// langVi is the locale MoMo renders its own pages and messages in for our merchants.
const langVi = "vi"

type createPaymentRequest struct {
	PartnerCode string `json:"partnerCode"`
	RequestId   string `json:"requestId"`
	Amount      int64  `json:"amount"`
	OrderId     string `json:"orderId"`
	OrderInfo   string `json:"orderInfo"`
	RedirectUrl string `json:"redirectUrl"`
	IpnUrl      string `json:"ipnUrl"`
	RequestType string `json:"requestType"`
	ExtraData   string `json:"extraData"`
	Lang        string `json:"lang"`
	Signature   string `json:"signature"`
}

type createPaymentResponse struct {
	PartnerCode  string `json:"partnerCode"`
	RequestId    string `json:"requestId"`
	OrderId      string `json:"orderId"`
	Amount       int64  `json:"amount"`
	ResponseTime int64  `json:"responseTime"`
	Message      string `json:"message"`
	ResultCode   int    `json:"resultCode"`
	PayUrl       string `json:"payUrl"`
	QrCodeUrl    string `json:"qrCodeUrl"`
	DeeplinkUrl  string `json:"deeplink"`
}

type refundRequest struct {
	PartnerCode string `json:"partnerCode"`
	OrderId     string `json:"orderId"`
	RequestId   string `json:"requestId"`
	Amount      int64  `json:"amount"`
	TransId     int64  `json:"transId"`
	Lang        string `json:"lang"`
	Description string `json:"description"`
	Signature   string `json:"signature"`
}

type refundResponse struct {
	PartnerCode  string `json:"partnerCode"`
	OrderId      string `json:"orderId"`
	RequestId    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	TransId      int64  `json:"transId"`
	ResultCode   int    `json:"resultCode"`
	Message      string `json:"message"`
	ResponseTime int64  `json:"responseTime"`
}

type queryRequest struct {
	PartnerCode string `json:"partnerCode"`
	RequestId   string `json:"requestId"`
	OrderId     string `json:"orderId"`
	Lang        string `json:"lang"`
	Signature   string `json:"signature"`
}

type queryResponse struct {
	PartnerCode  string `json:"partnerCode"`
	OrderId      string `json:"orderId"`
	RequestId    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	TransId      int64  `json:"transId"`
	ResultCode   int    `json:"resultCode"`
	Message      string `json:"message"`
	PayType      string `json:"payType"`
	ResponseTime int64  `json:"responseTime"`
}

// IpnPayload is the body MoMo posts to our callback when a payment settles. It is exported
// because the webhook handler decodes it before handing it here to be verified.
type IpnPayload struct {
	PartnerCode  string `json:"partnerCode"`
	OrderId      string `json:"orderId"`
	RequestId    string `json:"requestId"`
	Amount       int64  `json:"amount"`
	OrderInfo    string `json:"orderInfo"`
	OrderType    string `json:"orderType"`
	TransId      int64  `json:"transId"`
	ResultCode   int    `json:"resultCode"`
	Message      string `json:"message"`
	PayType      string `json:"payType"`
	ResponseTime int64  `json:"responseTime"`
	ExtraData    string `json:"extraData"`
	Signature    string `json:"signature"`
}
