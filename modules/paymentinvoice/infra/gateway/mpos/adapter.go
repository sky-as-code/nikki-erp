// Package mpos talks to mPOS (NextPay), which drives a physical card terminal.
//
// It is unlike the wallet gateways in two ways that shape everything here. First, it needs to be
// told *which terminal* to push the prompt to, so an order that names no terminal is meaningless
// — this is the adapter that makes real use of the port's ValidateOrder and PrepareMetadata.
// Second, it returns no QR code and no pay URL: the customer is standing at the terminal, and the
// prompt appears on the device rather than on a screen we control.
//
// Every request body is AES-128-ECB encrypted under the merchant secret; see crypto.go.
package mpos

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// Config is what this adapter needs to reach one mPOS merchant account.
type Config struct {
	MerchantId string

	// SecretKey must be exactly 16 bytes: it is used directly as the AES-128 key.
	SecretKey string
}

// Adapter implements itGateway.PaymentGateway for mPOS.
type Adapter struct {
	config Config
	caller *httpclient.HttpCaller
}

func NewAdapter(config Config, caller *httpclient.HttpCaller) *Adapter {
	return &Adapter{config: config, caller: caller}
}

func (this *Adapter) AdapterCode() string {
	return models.AdapterCodeMpos
}

// ValidateOrder rejects an order that does not say which terminal to charge.
//
// The gateway would reject it too, but only after the order had been recorded and the payment
// attempted; failing here means the caller is told which field is missing, and nothing is written.
func (this *Adapter) ValidateOrder(
	_ corectx.Context, req itGateway.OrderRequest, vErrs *ft.ClientErrors,
) error {
	if posIdOf(req.Metadata) == "" {
		vErrs.Append(*ft.NewBusinessViolation(
			models.OrderFieldMetadata+"."+models.OrderMetaPosId,
			"paymentinvoice.pos_id_required",
			"paying through a card terminal requires the terminal's id"))
	}
	return nil
}

// PrepareMetadata keeps the terminal id on the order.
//
// It is what the watchdog later needs in order to ask the gateway about an order that never
// received a callback, and what identifies the terminal whose queued orders are cleared.
func (this *Adapter) PrepareMetadata(
	_ corectx.Context, req itGateway.OrderRequest,
) (map[string]any, error) {
	return map[string]any{models.OrderMetaPosId: posIdOf(req.Metadata)}, nil
}

// CreatePayment pushes the order to the terminal, which then prompts the customer.
//
// It returns no QR code and no pay URL, and that is not an omission: the terminal displays the
// prompt itself.
func (this *Adapter) CreatePayment(
	ctx corectx.Context, req itGateway.CreatePaymentRequest,
) (*itGateway.CreatePaymentResult, error) {
	amount, err := toWholeUnits(req.Amount)
	if err != nil {
		return nil, err
	}

	config, err := this.resolveConfig(req.ProfileConfig)
	if err != nil {
		return nil, err
	}

	payload := addOrderRequest{
		ServiceName:   ServiceAddOrder,
		OrderId:       req.OrderCode,
		PosId:         posIdOf(req.Metadata),
		Amount:        amount,
		Description:   contentOf(req.Content, ""),
		PaymentType:   paymentTypeOther,
		PaymentMethod: PaymentMethodCard,
	}

	raw, err := this.call(ctx, config, pathOrder, payload, nil)
	if err != nil {
		return nil, err
	}

	return &itGateway.CreatePaymentResult{
		QrCodeUrl:   "",
		PayUrl:      "",
		RawResponse: raw,
	}, nil
}

func (this *Adapter) Refund(
	ctx corectx.Context, req itGateway.RefundRequest,
) (*itGateway.RefundResult, error) {
	amount, err := toWholeUnits(req.Amount)
	if err != nil {
		return nil, err
	}

	if req.RefTransactionId == "" {
		return nil, errors.New("mpos refund needs the gateway transaction code of the payment")
	}

	config, err := this.resolveConfig(req.ProfileConfig)
	if err != nil {
		return nil, err
	}

	payload := refundRequest{
		ServiceName:  ServiceRefundTransaction,
		TransCode:    req.RefTransactionId,
		RequestId:    uuid.NewString(),
		RefundAmount: amount,
	}

	var decoded refundResponse
	raw, err := this.call(ctx, config, pathTransaction, payload, &decoded)
	if err != nil {
		return nil, err
	}

	// The gateway echoes the transaction code of the refund, which may differ from the payment's.
	refTransactionId := decoded.TransCode
	if refTransactionId == "" {
		refTransactionId = req.RefTransactionId
	}

	return &itGateway.RefundResult{
		RefTransactionId: refTransactionId,
		RawResponse:      raw,
	}, nil
}

func (this *Adapter) CheckOrder(
	ctx corectx.Context, req itGateway.CheckOrderRequest,
) (*itGateway.CheckOrderResult, error) {
	posId := posIdOf(req.Metadata)
	if posId == "" {
		return nil, errors.New("mpos cannot be asked about an order that names no terminal")
	}

	amount, err := toWholeUnits(req.Amount)
	if err != nil {
		return nil, err
	}

	config, err := this.resolveConfig(req.ProfileConfig)
	if err != nil {
		return nil, err
	}

	payload := checkOrderRequest{
		ServiceName: ServiceGetTransactionStatus,
		OrderId:     req.OrderCode,
		PosId:       posId,
		// This request wants the amount as a string; the others want a number.
		Amount: strconv.FormatInt(amount, 10),
	}

	var decoded checkOrderResponse
	raw, err := this.call(ctx, config, pathOrder, payload, &decoded)
	if err != nil {
		return nil, err
	}

	return &itGateway.CheckOrderResult{
		Settled:          isSettled(decoded.TransStatus),
		Paid:             isPaid(decoded.TransStatus),
		RefTransactionId: decoded.TransCode,
		RawResponse:      raw,
	}, nil
}

// RemovePosOrders clears one order queued on a terminal.
//
// The caller decides which orders those are; this only speaks to the gateway. profileConfig is the
// credentials of the payment profile that queued the order, because a prompt can only be withdrawn
// by the merchant account that put it there.
func (this *Adapter) RemovePosOrders(
	ctx corectx.Context,
	orderCode string,
	posId string,
	amount decimal.Decimal,
	profileConfig map[string]any,
) error {
	wholeAmount, err := toWholeUnits(amount)
	if err != nil {
		return err
	}

	config, err := this.resolveConfig(profileConfig)
	if err != nil {
		return err
	}

	_, err = this.call(ctx, config, pathOrder, removeOrderRequest{
		ServiceName: ServiceRemoveOrder,
		OrderId:     orderCode,
		PosId:       posId,
		Amount:      wholeAmount,
	}, nil)
	return err
}

// DecryptWebhook turns the reqData of an inbound callback into its payload.
//
// Being able to decrypt it *is* the authentication: the body is encrypted under the merchant
// secret, which only the gateway and the account holder hold. A body that fails to decrypt did not
// come from mPOS under those credentials and must be refused rather than acted on.
//
// profileConfig is the credentials of the payment profile whose merchant id the callback names, or
// nil for the deployment's own account. It has to be resolved before this is called, because the
// callback carries nothing but that merchant id in the clear — the order it concerns is inside the
// very payload the credentials are needed to read.
func (this *Adapter) DecryptWebhook(
	reqData string, profileConfig map[string]any,
) (*WebhookPayload, error) {
	config, err := this.resolveConfig(profileConfig)
	if err != nil {
		return nil, err
	}

	var payload WebhookPayload
	if err := decrypt(reqData, config.SecretKey, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// MerchantId is the merchant account this deployment's own configuration names.
func (this *Adapter) MerchantId() string {
	return this.config.MerchantId
}

// WebhookOutcome reports what a callback means for the order it names.
func WebhookOutcome(payload WebhookPayload) (settled bool, paid bool) {
	return isSettled(payload.TransStatus), isPaid(payload.TransStatus)
}

// call encrypts a request, sends it, checks the gateway's result code and decrypts the reply.
//
// out may be nil for a request whose reply carries nothing worth decoding; the raw map is
// returned either way, so the caller can keep the gateway's own words as evidence.
func (this *Adapter) call(
	ctx corectx.Context, config Config, path string, payload any, out any,
) (map[string]any, error) {
	reqData, err := encrypt(payload, config.SecretKey)
	if err != nil {
		return nil, err
	}

	httpResponse, err := this.caller.Do(ctx, &httpclient.Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   envelope{MerchantId: config.MerchantId, ReqData: reqData},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "mpos request to %s failed", path)
	}

	var decoded response
	if err := json.Unmarshal(httpResponse.Body, &decoded); err != nil {
		return nil, errors.Wrapf(err, "mpos reply to %s could not be decoded", path)
	}

	if decoded.ResCode != ResponseCodeSuccess {
		return nil, errors.Errorf("mpos refused the request to %s: result code %d, %s",
			path, decoded.ResCode, decoded.Message)
	}

	raw := map[string]any{
		"resCode": decoded.ResCode,
		"message": decoded.Message,
	}

	if decoded.ResData == "" {
		return raw, nil
	}

	// Decode the inner payload twice: once into the caller's struct, once into a map kept as
	// evidence. A reply we cannot re-decode as a map is still one we understood, so losing the
	// evidence copy must not fail the payment.
	if out != nil {
		if err := decrypt(decoded.ResData, config.SecretKey, out); err != nil {
			return nil, err
		}
	}

	inner := map[string]any{}
	if err := decrypt(decoded.ResData, config.SecretKey, &inner); err == nil {
		raw["resData"] = inner
	}

	return raw, nil
}

// toWholeUnits converts an amount to the whole-unit integer mPOS expects.
//
// A fractional amount is refused rather than rounded: rounding would charge the customer standing
// at the terminal something other than what the order says.
func toWholeUnits(amount decimal.Decimal) (int64, error) {
	if !amount.Equal(amount.Truncate(0)) {
		return 0, errors.Errorf("mpos accepts whole amounts only, got %s", amount.String())
	}
	return amount.IntPart(), nil
}

// posIdOf reads the terminal id out of an order's metadata, tolerating its absence: the callers
// that require it check for the empty string and report it as a client error.
func posIdOf(metadata map[string]any) string {
	if metadata == nil {
		return ""
	}
	posId, _ := metadata[models.OrderMetaPosId].(string)
	return posId
}

func contentOf(content *string, fallback string) string {
	if content != nil && *content != "" {
		return *content
	}
	return fallback
}
