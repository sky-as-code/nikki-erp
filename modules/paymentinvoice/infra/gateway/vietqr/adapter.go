// Package vietqr talks to VietQR, which settles bank transfers made by scanning a QR code.
//
// It is the odd one out among the three gateways, in two ways worth knowing before reading on.
//
// First, the integration runs in both directions. We call VietQR to mint QR codes and to ask
// about transfers, authenticating with a bearer token; and the bank calls us when a transfer
// lands, authenticating with a bearer token that we issue. Those are two separate credential
// pairs, and conflating them is the mistake this package is arranged to prevent: the outbound
// pair lives in Config, and the inbound pair belongs to the webhook layer.
//
// Second, its reply to that inbound call is non-standard — the outcome is in the body, not the
// status code, and the bank keys off a three-digit `errorReason`. See WebhookResponse in wire.go.
package vietqr

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// qrContentMaxLength is the gateway's cap on the transfer description. Longer text is rejected
// outright rather than truncated, so the adapter trims it.
const qrContentMaxLength = 23

// Config is what this adapter needs to call VietQR. It holds the **outbound** credentials only;
// the pair the bank presents to us is the webhook layer's concern.
type Config struct {
	// Username and Password authenticate us to VietQR.
	Username string
	Password string

	// SecretKey is the shared secret the refund checksum is built on.
	SecretKey string

	BankCode   string
	BankNumber string
	BankName   string
}

// Adapter implements itGateway.PaymentGateway for VietQR.
type Adapter struct {
	config Config
	caller *httpclient.HttpCaller
	tokens *tokenCache
}

func NewAdapter(config Config, caller *httpclient.HttpCaller) *Adapter {
	return &Adapter{config: config, caller: caller, tokens: newTokenCache(nil)}
}

func (this *Adapter) AdapterCode() string {
	return models.AdapterCodeVietQr
}

// ValidateOrder has nothing to add: a bank transfer needs no input from the merchant beyond the
// amount, and the payer identifies themselves to their own bank.
func (this *Adapter) ValidateOrder(_ corectx.Context, _ itGateway.OrderRequest, _ *ft.ClientErrors) error {
	return nil
}

// PrepareMetadata returns nothing to store: VietQR needs no order-time input of its own.
func (this *Adapter) PrepareMetadata(_ corectx.Context, _ itGateway.OrderRequest) (map[string]any, error) {
	return nil, nil
}

func (this *Adapter) CreatePayment(
	ctx corectx.Context, req itGateway.CreatePaymentRequest,
) (*itGateway.CreatePaymentResult, error) {
	amount, err := toWholeUnits(req.Amount)
	if err != nil {
		return nil, err
	}

	payload := generateQrRequest{
		BankCode:     this.config.BankCode,
		BankAccount:  this.config.BankNumber,
		UserBankName: this.config.BankName,
		Content:      trimContent(contentOf(req.Content, req.OrderCode)),
		QrType:       QrTypeDynamic,
		Amount:       amount,
		OrderId:      req.OrderCode,
		TransType:    transTypeCredit,
	}

	var decoded generateQrResponse
	raw, err := this.post(ctx, pathGenerateQr, payload, &decoded)
	if err != nil {
		return nil, err
	}

	if decoded.QrCode == "" {
		return nil, errors.New("vietqr returned no QR code")
	}

	return &itGateway.CreatePaymentResult{
		QrCodeUrl:        decoded.QrCode,
		PayUrl:           decoded.QrLink,
		RefTransactionId: decoded.TransactionRefId,
		RawResponse:      raw,
	}, nil
}

func (this *Adapter) Refund(
	ctx corectx.Context, req itGateway.RefundRequest,
) (*itGateway.RefundResult, error) {
	if req.RefTransactionId == "" {
		return nil, errors.New("vietqr refund needs the reference number of the transfer")
	}

	amount, err := toWholeUnits(req.Amount)
	if err != nil {
		return nil, err
	}

	// The amount string is sent and checksummed, so it is rendered once and used for both. The
	// two disagreeing is a rejected refund with no explanation.
	amountText := strconv.FormatInt(amount, 10)

	payload := refundRequest{
		BankAccount:     this.config.BankNumber,
		ReferenceNumber: req.RefTransactionId,
		Amount:          amountText,
		Content:         contentOf(req.Content, "Refund "+req.OrderCode),
		CheckSum: refundChecksum(
			this.config.SecretKey, req.RefTransactionId, amountText, this.config.BankNumber),
		BankCode: this.config.BankCode,
	}

	var decoded refundResponse
	raw, err := this.post(ctx, pathRefund, payload, &decoded)
	if err != nil {
		return nil, err
	}

	if decoded.Status != RefundStatusSuccess {
		return nil, errors.Errorf("vietqr refused the refund: %s, %s", decoded.Status, decoded.Message)
	}

	return &itGateway.RefundResult{
		// VietQR issues no new identifier for a refund; it is filed against the original
		// transfer, so that is what stays recorded against it.
		RefTransactionId: req.RefTransactionId,
		RawResponse:      raw,
	}, nil
}

func (this *Adapter) CheckOrder(
	ctx corectx.Context, req itGateway.CheckOrderRequest,
) (*itGateway.CheckOrderResult, error) {
	payload := checkOrderRequest{
		BankAccount: this.config.BankNumber,
		Type:        checkOrderTypeByOrderId,
		Value:       req.OrderCode,
		CheckSum:    checkOrderChecksum(this.config.BankNumber, this.config.Username),
	}

	// This reply is a JSON array rather than an object, so it is decoded on its own rather than
	// through post's struct path.
	body, err := this.send(ctx, pathCheckOrder, payload)
	if err != nil {
		return nil, err
	}

	var entries []checkOrderEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, errors.Wrap(err, "vietqr check-order reply could not be decoded")
	}

	raw := map[string]any{}
	_ = json.Unmarshal(body, &raw)

	// No entry at all means the gateway has never seen a transfer for this order: nothing has
	// been paid, but nothing has failed either, so it is left for the next sweep to expire.
	if len(entries) == 0 {
		return &itGateway.CheckOrderResult{Settled: false, RawResponse: raw}, nil
	}

	entry := entries[0]
	return &itGateway.CheckOrderResult{
		Settled:          entry.Status != CheckOrderStatusPending,
		Paid:             entry.Status == CheckOrderStatusPaid,
		RefTransactionId: entry.ReferenceNumber,
		RawResponse:      raw,
	}, nil
}

// post sends a JSON request under a bearer token and decodes the reply into out.
func (this *Adapter) post(
	ctx corectx.Context, path string, payload any, out any,
) (map[string]any, error) {
	body, err := this.send(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, out); err != nil {
		return nil, errors.Wrapf(err, "vietqr reply to %s could not be decoded", path)
	}

	raw := map[string]any{}
	// A reply we cannot re-decode as a map is still one we understood, so losing the evidence
	// copy must not fail the payment.
	_ = json.Unmarshal(body, &raw)
	return raw, nil
}

// send authenticates and sends one request, retrying once if the gateway rejects our token.
//
// The single retry exists because the gateway's idea of when a session ends is the one that
// counts: it may drop a token we still believe is live, and a payment should not fail for that.
func (this *Adapter) send(ctx corectx.Context, path string, payload any) ([]byte, error) {
	response, err := this.sendOnce(ctx, path, payload)
	if err == nil {
		return response, nil
	}

	if !errors.Is(err, errUnauthorized) {
		return nil, err
	}

	this.tokens.invalidate()
	return this.sendOnce(ctx, path, payload)
}

// errUnauthorized marks the one failure worth retrying: a token the gateway would not accept.
var errUnauthorized = errors.New("vietqr rejected the bearer token")

func (this *Adapter) sendOnce(ctx corectx.Context, path string, payload any) ([]byte, error) {
	token, err := this.tokens.get(func() (string, time.Duration, error) {
		return this.login(ctx)
	})
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Authorization", TokenTypeBearer+" "+token)

	response, err := this.caller.Do(ctx, &httpclient.Request{
		Method:  http.MethodPost,
		Path:    path,
		Headers: headers,
		Body:    payload,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "vietqr request to %s failed", path)
	}

	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return nil, errUnauthorized
	}

	return response.Body, nil
}

// login exchanges the outbound credentials for a bearer token.
//
// VietQR takes them as HTTP Basic, which is why they are encoded here rather than sent as a body.
func (this *Adapter) login(ctx corectx.Context) (string, time.Duration, error) {
	headers := http.Header{}
	headers.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString(
		[]byte(this.config.Username+":"+this.config.Password)))

	response, err := this.caller.Do(ctx, &httpclient.Request{
		Method:  http.MethodPost,
		Path:    pathTokenGenerate,
		Headers: headers,
	})
	if err != nil {
		return "", 0, errors.Wrap(err, "vietqr login failed")
	}

	var decoded TokenResponse
	if err := json.Unmarshal(response.Body, &decoded); err != nil {
		return "", 0, errors.Wrap(err, "vietqr login reply could not be decoded")
	}

	if decoded.AccessToken == "" {
		// Deliberately says nothing about the credentials themselves.
		return "", 0, errors.New("vietqr login returned no access token")
	}

	return decoded.AccessToken, time.Duration(decoded.ExpiresIn) * time.Second, nil
}

// toWholeUnits converts an amount to the whole-unit integer VietQR expects.
func toWholeUnits(amount decimal.Decimal) (int64, error) {
	if !amount.Equal(amount.Truncate(0)) {
		return 0, errors.Errorf("vietqr accepts whole amounts only, got %s", amount.String())
	}
	return amount.IntPart(), nil
}

// trimContent cuts the description to what the gateway accepts. Sending more is rejected rather
// than truncated at the far end, which would fail the payment over a long description.
func trimContent(content string) string {
	if len(content) <= qrContentMaxLength {
		return content
	}
	return content[:qrContentMaxLength]
}

func contentOf(content *string, fallback string) string {
	if content != nil && *content != "" {
		return *content
	}
	return fallback
}
