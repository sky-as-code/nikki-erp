package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Starting a gateway payment.
//
// A CASH TENDER AND A GATEWAY TENDER ARE THE SAME BUSINESS EVENT REACHED TWO WAYS. Cash is counted
// at the counter and is simply in, which is what RecordPayment does by default. A card or a QR code
// has to be collected: an order is opened with the provider, the customer is shown something to pay
// with, and the money arrives later — or never.
//
// So this writes the payment as `pending` and stops. It does not settle the bill, because nothing
// has been paid yet; the settlement event does that when the provider says so, and the
// reconciliation sweep covers the events that never arrive. Anything here that optimistically
// counted the money would settle a bill against funds that may still be declined.
//
// Every gate RecordPayment applies is applied here first, through RecordPayment itself: opening an
// order for a method the channel does not accept would cost a real gateway round trip to learn what
// Sales already knew.

// StartGatewayPaymentParams is what opening a collection needs.
type StartGatewayPaymentParams struct {
	SalesBillId     string
	PaymentMethodId string

	Amount       decimal.Decimal
	CurrencyCode string

	// Content is what the payer sees on their statement. Empty lets the provider fall back to the
	// order identifier, which is still traceable.
	Content string
}

// StartGatewayPaymentResult is what the till puts in front of the customer.
type StartGatewayPaymentResult struct {
	SalesPaymentId string
	SalesBillId    string

	PaymentOrderId string
	OrderCode      string

	// QrCodeUrl and PayUrl are both empty for a card terminal, where the prompt is pushed to the
	// device the customer is standing at.
	QrCodeUrl string
	PayUrl    string
}

// The refusal reasons opening a gateway collection can produce, beyond the ones RecordPayment
// already reports.
const (
	ReasonGatewayUnavailable   = "sales_payment.gateway_unavailable"
	ReasonMethodHasNoGateway   = "sales_payment.method_has_no_gateway"
	ReasonGatewayRefusedOrder  = "sales_payment.gateway_refused"
	ReasonGatewayOrderNotOpened = "sales_payment.gateway_order_not_opened"
)

// StartGatewayPayment opens a collection with the provider and records the payment awaiting it.
func StartGatewayPayment(
	ctx corectx.Context,
	params StartGatewayPaymentParams,
	methods itExt.PaymentMethodExtService,
	orders itExt.PaymentOrderExtService,
	channelPayments itChannel.ChannelPaymentAppService,
	policy SalesPolicy,
) (*StartGatewayPaymentResult, *ft.ClientErrors, error) {
	if orders == nil {
		// Default-deny, like every other unavailable port in this module: without the gateway there
		// is no way to collect, and pretending otherwise would record money that was never asked for.
		return nil, refusal("payment_method_id", ReasonGatewayUnavailable,
			"the payment gateway is unavailable, so a collection cannot be started"), nil
	}

	bill, err := loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SalesBillId)
	if err != nil {
		return nil, nil, err
	}
	if bill == nil {
		return nil, refusal("sales_bill_id", ReasonBillNotFound,
			"no bill exists with id '"+params.SalesBillId+"'"), nil
	}

	if vErrs, err := assertMethodCollectsThroughGateway(
		ctx, params, methods,
	); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// The payment is written first, and deliberately: it is the record that a collection was
	// attempted. Opening the order first and dying before the write would leave money collectable
	// against a bill with nothing awaiting it, which no sweep could then reconcile.
	recorded, vErrs, err := RecordPayment(ctx, RecordPaymentParams{
		SalesBillId:     params.SalesBillId,
		PaymentMethodId: params.PaymentMethodId,
		Amount:          params.Amount,
		CurrencyCode:    params.CurrencyCode,
		Status:          string(models.SalesPaymentStatusPending),
	}, methods, channelPayments, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	opened, err := orders.CreatePayment(ctx, itExt.CreateGatewayPaymentCommand{
		OrgId:           stringOf(bill, models.SalesPaymentFieldOrgId),
		PaymentMethodId: params.PaymentMethodId,
		Amount:          params.Amount,
		Content:         params.Content,
		SalesPaymentId:  recorded.SalesPaymentId,
		SalesBillId:     params.SalesBillId,
	})
	if err != nil {
		return nil, nil, err
	}

	if opened.Refused || !opened.HasData {
		// The gateway would not take it. The payment is failed rather than deleted: it is evidence
		// the attempt was made, and it frees the method slot so the customer can try another card.
		if err := failGatewayPayment(ctx, recorded.SalesPaymentId); err != nil {
			return nil, nil, err
		}
		reason := opened.RefusalReason
		if reason == "" {
			reason = "the payment provider did not open an order"
		}
		return nil, refusal("payment_method_id", ReasonGatewayRefusedOrder, reason), nil
	}

	if err := attachPaymentOrder(ctx, recorded.SalesPaymentId, opened.Data); err != nil {
		return nil, nil, err
	}

	return &StartGatewayPaymentResult{
		SalesPaymentId: recorded.SalesPaymentId,
		SalesBillId:    params.SalesBillId,
		PaymentOrderId: opened.Data.OrderId,
		OrderCode:      opened.Data.OrderCode,
		QrCodeUrl:      opened.Data.QrCodeUrl,
		PayUrl:         opened.Data.PayUrl,
	}, nil, nil
}

// assertMethodCollectsThroughGateway refuses a method that takes money at the counter.
//
// Sales asks whether the method has a gateway at all, never which one: that is paymentinvoice's
// business, and branching on momo or vietqr here would put a payment integration Sales does not own
// into Sales. Cash sent down this path would open an order nobody can pay.
func assertMethodCollectsThroughGateway(
	ctx corectx.Context,
	params StartGatewayPaymentParams,
	methods itExt.PaymentMethodExtService,
) (*ft.ClientErrors, error) {
	if methods == nil {
		return refusal("payment_method_id", ReasonMethodNotUsable,
			"the payment method service is unavailable, so usability cannot be confirmed"), nil
	}

	amount := params.Amount
	usable, err := methods.AssertUsable(ctx, itExt.AssertUsableQuery{
		PaymentMethodId: params.PaymentMethodId,
		Amount:          &amount,
	})
	if err != nil {
		return nil, err
	}
	if usable == nil || !usable.HasData {
		return refusal("payment_method_id", ReasonMethodNotUsable,
			"this payment method cannot currently take a payment"), nil
	}
	if !usable.Data.HasGateway {
		return refusal("payment_method_id", ReasonMethodHasNoGateway,
			"payment method '"+usable.Data.Code+"' is settled at the counter, "+
				"so it is recorded directly rather than collected through a gateway"), nil
	}
	return nil, nil
}

// attachPaymentOrder stores the correlation a settlement arrives on.
func attachPaymentOrder(
	ctx corectx.Context, salesPaymentId string, opened itExt.CreateGatewayPaymentResultData,
) error {
	fields := dmodel.DynamicFields{
		models.SalesPaymentFieldPaymentOrderId: opened.OrderId,
	}
	if opened.OrderCode != "" {
		// The gateway's own key, kept for reconciliation and support. provider_reference rather than
		// external_transaction_id: that one means the id issued when money moves, and writing a code
		// there now would claim a settled transaction and trip the replay guard.
		fields[models.SalesPaymentFieldProviderReference] = opened.OrderCode
	}
	return updatePaymentFields(ctx, salesPaymentId, fields)
}

// failGatewayPayment closes a payment whose collection never started.
func failGatewayPayment(ctx corectx.Context, salesPaymentId string) error {
	return updatePaymentFields(ctx, salesPaymentId, dmodel.DynamicFields{
		models.SalesPaymentFieldStatus: string(models.SalesPaymentStatusFailed),
	})
}

// updatePaymentFields re-reads the payment before writing, because writeChanges carries the row's
// etag into the update so that a concurrent writer loses rather than overwrites. RecordPayment hands
// back an id, not the row.
func updatePaymentFields(
	ctx corectx.Context, salesPaymentId string, fields dmodel.DynamicFields,
) error {
	payment, err := loadRecord(ctx,
		models.SalesPaymentSchemaName, models.SalesPaymentFieldId, salesPaymentId)
	if err != nil {
		return err
	}
	if payment == nil {
		return nil
	}
	return writeChanges(ctx, models.SalesPaymentSchemaName, payment, fields)
}

// refusal builds a one-violation refusal, the shape every gate in this file returns.
func refusal(field, reason, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
	return vErrs
}
