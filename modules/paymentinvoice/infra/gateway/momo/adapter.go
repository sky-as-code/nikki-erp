// Package momo talks to MoMo, a Vietnamese e-wallet.
//
// MoMo settles asynchronously: creating a payment returns a QR code and a pay URL for the payer,
// and the outcome arrives later as an IPN callback signed with the merchant secret. Nothing in
// the create response says the payment succeeded, so nothing outside this package should treat it
// as though it had.
//
// The wire contract — the field sets that are signed, their ordering, and the shape of each
// request — is MoMo's and is transcribed from the NestJS service this module supersedes. See
// signature.go for why none of it may be tidied.
package momo

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

// Config is what this adapter needs to reach one MoMo merchant account.
type Config struct {
	PartnerCode string
	AccessKey   string
	SecretKey   string
	ApiEndpoint string

	// IpnUrl is where MoMo posts the result. It is sent on every create, so changing it needs
	// nothing done on MoMo's side.
	IpnUrl string

	// RedirectUrl is where the payer's browser lands afterwards. A human destination: nothing
	// is settled by a visit to it.
	RedirectUrl string
}

// Adapter implements itGateway.PaymentGateway for MoMo.
type Adapter struct {
	config Config
	caller *httpclient.HttpCaller
}

func NewAdapter(config Config, caller *httpclient.HttpCaller) *Adapter {
	return &Adapter{config: config, caller: caller}
}

func (this *Adapter) AdapterCode() string {
	return models.AdapterCodeMomo
}

// ValidateOrder has nothing to add beyond what the order service already checks.
//
// A wallet payment needs no input from the merchant: the payer identifies themselves to MoMo, not
// to us. The per-method amount bounds live on the payment-method row, where an administrator can
// adjust them when the contract changes, rather than being compiled in here.
func (this *Adapter) ValidateOrder(_ corectx.Context, _ itGateway.OrderRequest, _ *ft.ClientErrors) error {
	return nil
}

// PrepareMetadata returns nothing to store: MoMo needs no order-time input of its own.
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

	// MoMo knows the order by its order_code, never by our id or our quoted order_id: the code is
	// what its callback will arrive under.
	payload := createPaymentRequest{
		PartnerCode: this.config.PartnerCode,
		RequestId:   uuid.NewString(),
		Amount:      amount,
		OrderId:     req.OrderCode,
		OrderInfo:   orderInfoOf(req),
		RedirectUrl: this.config.RedirectUrl,
		IpnUrl:      this.config.IpnUrl,
		RequestType: requestTypeCaptureWallet,
		ExtraData:   "",
		Lang:        langVi,
	}
	payload.Signature = signingFields{
		"accessKey":   this.config.AccessKey,
		"amount":      strconv.FormatInt(payload.Amount, 10),
		"extraData":   payload.ExtraData,
		"ipnUrl":      payload.IpnUrl,
		"orderId":     payload.OrderId,
		"orderInfo":   payload.OrderInfo,
		"partnerCode": payload.PartnerCode,
		"redirectUrl": payload.RedirectUrl,
		"requestId":   payload.RequestId,
		"requestType": payload.RequestType,
	}.sign(this.config.SecretKey)

	var response createPaymentResponse
	raw, err := this.post(ctx, "/create", payload, &response)
	if err != nil {
		return nil, err
	}

	if response.ResultCode != ResultCodeSuccess {
		return nil, errors.Errorf("momo refused to create the payment: result code %d, %s",
			response.ResultCode, response.Message)
	}

	return &itGateway.CreatePaymentResult{
		QrCodeUrl:   response.QrCodeUrl,
		PayUrl:      response.PayUrl,
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

	// MoMo identifies the payment being reversed by its own numeric transaction id, which we
	// recorded when the payment settled.
	transId, err := strconv.ParseInt(req.RefTransactionId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err,
			"momo refund needs the gateway transaction id, got %q", req.RefTransactionId)
	}

	payload := refundRequest{
		PartnerCode: this.config.PartnerCode,
		// A refund is its own operation to MoMo and needs an id distinct from the payment's.
		OrderId:     uuid.NewString(),
		RequestId:   uuid.NewString(),
		Amount:      amount,
		TransId:     transId,
		Lang:        langVi,
		Description: descriptionOf(req),
	}
	payload.Signature = signingFields{
		"accessKey":   this.config.AccessKey,
		"amount":      strconv.FormatInt(payload.Amount, 10),
		"description": payload.Description,
		"orderId":     payload.OrderId,
		"partnerCode": payload.PartnerCode,
		"requestId":   payload.RequestId,
		"transId":     strconv.FormatInt(payload.TransId, 10),
	}.sign(this.config.SecretKey)

	var response refundResponse
	raw, err := this.post(ctx, "/refund", payload, &response)
	if err != nil {
		return nil, err
	}

	if response.ResultCode != ResultCodeSuccess {
		return nil, errors.Errorf("momo refused the refund: result code %d, %s",
			response.ResultCode, response.Message)
	}

	return &itGateway.RefundResult{
		RefTransactionId: strconv.FormatInt(response.TransId, 10),
		RawResponse:      raw,
	}, nil
}

func (this *Adapter) CheckOrder(
	ctx corectx.Context, req itGateway.CheckOrderRequest,
) (*itGateway.CheckOrderResult, error) {
	payload := queryRequest{
		PartnerCode: this.config.PartnerCode,
		RequestId:   uuid.NewString(),
		OrderId:     req.OrderCode,
		Lang:        langVi,
	}
	payload.Signature = signingFields{
		"accessKey":   this.config.AccessKey,
		"orderId":     payload.OrderId,
		"partnerCode": payload.PartnerCode,
		"requestId":   payload.RequestId,
	}.sign(this.config.SecretKey)

	var response queryResponse
	raw, err := this.post(ctx, "/query", payload, &response)
	if err != nil {
		return nil, err
	}

	// A payment MoMo is still working on is not a verdict. Reporting it as unsettled leaves the
	// order alone for the next sweep; reporting it as unpaid would expire an order that is about
	// to succeed.
	settled := response.ResultCode != ResultCodeTransPending &&
		response.ResultCode != ResultCodeTransProcessing

	return &itGateway.CheckOrderResult{
		Settled:          settled,
		Paid:             response.ResultCode == ResultCodeSuccess,
		RefTransactionId: strconv.FormatInt(response.TransId, 10),
		RawResponse:      raw,
	}, nil
}

// VerifyIpn reports whether an IPN callback really came from MoMo.
//
// The thirteen fields below are MoMo's IPN signing set, and it is not the same set as any of the
// request signatures — notably it includes message, orderType, payType, responseTime and
// resultCode, which no request signs.
func (this *Adapter) VerifyIpn(payload IpnPayload) bool {
	return signingFields{
		"accessKey":    this.config.AccessKey,
		"amount":       strconv.FormatInt(payload.Amount, 10),
		"extraData":    payload.ExtraData,
		"message":      payload.Message,
		"orderId":      payload.OrderId,
		"orderInfo":    payload.OrderInfo,
		"orderType":    payload.OrderType,
		"partnerCode":  payload.PartnerCode,
		"payType":      payload.PayType,
		"requestId":    payload.RequestId,
		"responseTime": strconv.FormatInt(payload.ResponseTime, 10),
		"resultCode":   strconv.Itoa(payload.ResultCode),
		"transId":      strconv.FormatInt(payload.TransId, 10),
	}.verify(this.config.SecretKey, payload.Signature)
}

// post sends one JSON request and decodes the reply into out, also returning it as a raw map so
// that the caller can keep the gateway's own words as evidence.
func (this *Adapter) post(
	ctx corectx.Context, path string, body any, out any,
) (map[string]any, error) {
	response, err := this.caller.Do(ctx, &httpclient.Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   body,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "momo request to %s failed", path)
	}

	if err := json.Unmarshal(response.Body, out); err != nil {
		return nil, errors.Wrapf(err, "momo reply to %s could not be decoded", path)
	}

	raw := map[string]any{}
	// A reply we cannot re-decode as a map is still a reply we understood; losing the evidence
	// copy must not fail the payment.
	_ = json.Unmarshal(response.Body, &raw)
	return raw, nil
}

// toWholeUnits converts an amount to the whole-unit integer MoMo expects.
//
// MoMo deals only in whole dong. Amounts are held as decimal because how many minor units a
// currency has is a property of the currency, so the conversion happens here, at the boundary
// with the one gateway that requires it, and a fractional amount is refused rather than rounded
// into a figure the payer did not agree to.
func toWholeUnits(amount decimal.Decimal) (int64, error) {
	if !amount.Equal(amount.Truncate(0)) {
		return 0, errors.Errorf("momo accepts whole amounts only, got %s", amount.String())
	}
	return amount.IntPart(), nil
}

func orderInfoOf(req itGateway.CreatePaymentRequest) string {
	if req.Content != nil && *req.Content != "" {
		return *req.Content
	}
	return req.OrderCode
}

func descriptionOf(req itGateway.RefundRequest) string {
	if req.Content != nil && *req.Content != "" {
		return *req.Content
	}
	return "Refund for " + req.OrderCode
}
