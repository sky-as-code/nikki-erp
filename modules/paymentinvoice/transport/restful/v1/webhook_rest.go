// Package v1 serves the gateway callbacks.
//
// These are the one part of the module that cannot be an engine action. They are called by MoMo,
// by NextPay and by the bank — each of which has a fixed idea of the request and the response, and
// none of which can present a user's authorization. So they are hand-written, registered on the
// public route group, and each authenticates its caller in its own way:
//
//   - MoMo signs its callback with the merchant secret, over a fixed 13-field set.
//   - mPOS encrypts the whole body under the merchant secret; being able to decrypt it is the
//     authentication.
//   - VietQR has this deployment host a token endpoint the bank logs into, and then presents the
//     bearer we issued.
//
// A body that fails its check is refused without touching an order. None of these endpoints tells
// an unauthenticated caller whether an order exists: that would let anyone enumerate order codes.
package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/momo"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/mpos"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/vietqr"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// InboundAuth handles the credentials and tokens the bank uses to authenticate to us.
type InboundAuth interface {
	ValidateBasic(username, password string) bool
	IssueToken(username string) (token string, expiresIn int64, err error)
	ValidateBearer(authorization string) bool
}

// WebhookRest serves the three gateways' callbacks.
type WebhookRest struct {
	orders   *services.OrderDomainService
	registry *itGateway.Registry
	inbound  InboundAuth
	logger   logging.LoggerService
}

func NewWebhookRest(
	orders *services.OrderDomainService,
	registry *itGateway.Registry,
	inbound InboundAuth,
	logger logging.LoggerService,
) *WebhookRest {
	return &WebhookRest{orders: orders, registry: registry, inbound: inbound, logger: logger}
}

// MomoIpn receives MoMo's payment result.
//
// MoMo expects 204 whatever the outcome, and retries anything else. That includes the cases below
// where nothing is applied: a callback for an unknown order is not something MoMo can fix by
// sending it again, and a signature that does not verify did not come from MoMo at all.
func (this *WebhookRest) MomoIpn(echoCtx *echo.Context) error {
	var payload momo.IpnPayload
	if err := echoCtx.Bind(&payload); err != nil {
		this.logger.Warnf("paymentinvoice: malformed MoMo IPN body: %s", err.Error())
		return echoCtx.NoContent(http.StatusNoContent)
	}

	adapter, ok := adapterAs[*momo.Adapter](this.registry, models.AdapterCodeMomo)
	if !ok {
		// MoMo is calling a deployment that does not have it enabled. Nothing here can verify the
		// signature, so nothing here may act on the body.
		this.logger.Warnf("paymentinvoice: a MoMo IPN arrived but the MoMo gateway is not enabled")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	reqCtx, ok := requestContext(echoCtx)
	if !ok {
		this.logger.Errorf("paymentinvoice: a webhook ran without a request context")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	// The signature is MoMo's, made with the secret of the account that took the money — so the
	// order has to be found before the callback can be checked, to know which account that was.
	// An unknown order code yields no credentials and the check then fails, which is the same
	// answer an unknown code got before: a 204 that says nothing either way.
	profileConfig, err := this.orders.ProfileConfigForOrderCode(reqCtx, payload.OrderId)
	if err != nil {
		this.logger.Errorf("paymentinvoice: resolving a MoMo IPN's payment profile failed: %s", err.Error())
		return echoCtx.NoContent(http.StatusNoContent)
	}

	if !adapter.VerifyIpn(payload, profileConfig) {
		// Deliberately not distinguished from a valid callback in the response: telling an
		// unsigned caller that their signature was wrong helps them work out a correct one.
		this.logger.Warnf("paymentinvoice: a MoMo IPN failed signature verification")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	this.apply(echoCtx, services.GatewayResult{
		OrderCode:        payload.OrderId,
		Paid:             payload.ResultCode == momo.ResultCodeSuccess,
		RefTransactionId: strconv.FormatInt(payload.TransId, 10),
		RefPayload:       asPayloadMap(payload),
	})
	return echoCtx.NoContent(http.StatusNoContent)
}

// MposWebhook receives the card terminal's result.
//
// The body is an encrypted envelope, and being able to decrypt it is the authentication: it is
// encrypted under the merchant secret, which only the gateway and this deployment hold.
func (this *WebhookRest) MposWebhook(echoCtx *echo.Context) error {
	var envelope mpos.WebhookEnvelope
	if err := echoCtx.Bind(&envelope); err != nil {
		return echoCtx.NoContent(http.StatusNoContent)
	}

	adapter, ok := adapterAs[*mpos.Adapter](this.registry, models.AdapterCodeMpos)
	if !ok {
		this.logger.Warnf("paymentinvoice: an mPOS callback arrived but the gateway is not enabled")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	reqCtx, ok := requestContext(echoCtx)
	if !ok {
		this.logger.Errorf("paymentinvoice: a webhook ran without a request context")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	profileConfig, found := this.mposProfileConfig(reqCtx, adapter, envelope.MerchantId)
	if !found {
		// The callback names a merchant account this deployment holds no credentials for, so there
		// is nothing here that could read the body, let alone act on it.
		this.logger.Warnf("paymentinvoice: an mPOS callback named an unknown merchant account")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	payload, err := adapter.DecryptWebhook(envelope.ReqData, profileConfig)
	if err != nil {
		// A body that will not decrypt did not come from mPOS, so it is refused rather than acted
		// on. The reason is logged but not returned, for the same reason as MoMo's signature.
		this.logger.Warnf("paymentinvoice: an mPOS callback could not be decrypted")
		return echoCtx.NoContent(http.StatusNoContent)
	}

	_, paid := mpos.WebhookOutcome(*payload)
	this.apply(echoCtx, services.GatewayResult{
		OrderCode:        payload.OrderId,
		Paid:             paid,
		RefTransactionId: payload.TransCode,
		RefPayload:       asPayloadMap(payload),
	})
	return echoCtx.NoContent(http.StatusNoContent)
}

// VietQrTokenGenerate issues the bearer the bank presents on its transaction callbacks.
//
// VietQR's integration is unusual in having the partner host this: the bank logs in here with
// HTTP Basic and gets back a short-lived token of our own issuing. These are not the credentials
// this deployment uses when it calls VietQR — that is a separate pair, and conflating them would
// have us present the bank's own credentials back to it.
func (this *WebhookRest) VietQrTokenGenerate(echoCtx *echo.Context) error {
	username, password, ok := echoCtx.Request().BasicAuth()
	if !ok || !this.inbound.ValidateBasic(username, password) {
		return echoCtx.JSON(http.StatusUnauthorized, map[string]any{
			"error":       true,
			"errorReason": vietqr.WebhookReasonNotFound,
		})
	}

	token, expiresIn, err := this.inbound.IssueToken(username)
	if err != nil {
		return err
	}

	return echoCtx.JSON(http.StatusOK, vietqr.TokenResponse{
		AccessToken: token,
		TokenType:   vietqr.TokenTypeBearer,
		ExpiresIn:   expiresIn,
	})
}

// VietQrTransactionSync receives a settled bank transfer.
//
// Its reply is non-standard and must stay byte-exact: the bank reads the body rather than the
// status code, and keys off errorReason. That is why the two replies are only ever built by
// vietqr.NewWebhookAccepted and vietqr.NewWebhookNotFound.
func (this *WebhookRest) VietQrTransactionSync(echoCtx *echo.Context) error {
	if !this.inbound.ValidateBearer(echoCtx.Request().Header.Get("Authorization")) {
		return echoCtx.JSON(http.StatusUnauthorized, vietqr.NewWebhookNotFound())
	}

	var payload vietqr.WebhookPayload
	if err := echoCtx.Bind(&payload); err != nil {
		return echoCtx.JSON(http.StatusBadRequest, vietqr.NewWebhookNotFound())
	}

	reqCtx, ok := requestContext(echoCtx)
	if !ok {
		return echoCtx.JSON(http.StatusInternalServerError, vietqr.NewWebhookNotFound())
	}

	// A bank transfer is only ever reported once it has arrived, so this callback existing is
	// itself the confirmation of payment. There is no failure form of it.
	outcome, err := this.orders.ApplyGatewayResult(reqCtx, services.GatewayResult{
		OrderCode:        payload.OrderId,
		Paid:             true,
		RefTransactionId: payload.ReferenceNumber,
		RefPayload:       asPayloadMap(payload),
	})
	if err != nil {
		return err
	}

	if !outcome.OrderFound {
		return echoCtx.JSON(http.StatusOK, vietqr.NewWebhookNotFound())
	}
	// An order already settled answers success rather than "not found": the bank retries anything
	// else, and the transfer really has been accounted for.
	return echoCtx.JSON(http.StatusOK, vietqr.NewWebhookAccepted(payload.ReferenceNumber))
}

// apply records a gateway's verdict, logging rather than failing when nothing could be applied.
//
// MoMo and mPOS are told nothing either way — both expect an empty 204 — so an unknown order or a
// replayed callback has to be visible in the log or it is visible nowhere.
func (this *WebhookRest) apply(echoCtx *echo.Context, result services.GatewayResult) {
	reqCtx, ok := requestContext(echoCtx)
	if !ok {
		this.logger.Errorf("paymentinvoice: a webhook ran without a request context")
		return
	}

	outcome, err := this.orders.ApplyGatewayResult(reqCtx, result)
	if err != nil {
		this.logger.Errorf("paymentinvoice: applying a gateway result failed: %s", err.Error())
		return
	}
	if !outcome.OrderFound {
		this.logger.Warnf("paymentinvoice: a callback named unknown order code '%s'", result.OrderCode)
		return
	}
	if !outcome.Applied {
		// Ordinary: every one of these gateways retries a callback it did not get a clean answer
		// to, so the same result arriving twice is expected rather than exceptional.
		this.logger.Infof("paymentinvoice: order '%s' had already settled; callback ignored", outcome.OrderId)
	}
}

// adapterAs fetches a registered adapter and narrows it to its concrete type.
//
// The webhook handling needs methods that are deliberately not on the gateway port — verifying a
// signature and decrypting an envelope are things only the gateway that defined them can do, and
// putting them on the port would make every other adapter stub them.
func adapterAs[T any](registry *itGateway.Registry, adapterCode string) (T, bool) {
	var zero T
	adapter, exists := registry.Get(adapterCode)
	if !exists {
		return zero, false
	}
	typed, ok := adapter.(T)
	return typed, ok
}

// asPayloadMap renders a callback body as the map the transaction's ref_payload column holds.
//
// It goes through JSON so the stored evidence carries the gateway's own field names, which is
// what someone reconciling a disputed payment will be comparing against.
func asPayloadMap(payload any) map[string]any {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil
	}
	return decoded
}

// requestContext narrows the request's context to the module's own.
func requestContext(echoCtx *echo.Context) (corectx.Context, bool) {
	reqCtx, ok := echoCtx.Request().Context().(corectx.Context)
	return reqCtx, ok
}

// mposProfileConfig finds the credentials that can read one inbound card-terminal callback.
//
// The callback carries the merchant account in the clear and everything else encrypted under that
// account's secret, so the account has to be identified before the body can be read at all. The
// deployment's own account is tried first because it is the common case and costs no query; only
// then are the payment profiles scanned.
//
// It reports false when no account matches. That is not the same as a body that fails to decrypt:
// here nothing was even attempted, and attempting it under the wrong secret would produce a
// decrypt failure that reads as a forged callback rather than as a missing profile.
func (this *WebhookRest) mposProfileConfig(
	ctx corectx.Context, adapter *mpos.Adapter, merchantId string,
) (map[string]any, bool) {
	// A callback that names no merchant, or names this deployment's own, is served by the
	// configured credentials — which is every callback on a deployment that has no profiles.
	if merchantId == "" || merchantId == adapter.MerchantId() {
		return nil, true
	}

	configs, err := this.orders.ProfileConfigsByMethod(ctx, models.PaymentProfileMethodMpos)
	if err != nil {
		this.logger.Errorf("paymentinvoice: reading the mPOS payment profiles failed: %s", err.Error())
		return nil, false
	}

	for _, config := range configs {
		if mpos.MerchantIdOf(config) == merchantId {
			return config, true
		}
	}
	return nil, false
}
