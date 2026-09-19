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

	// Amount is optional: zero means the whole of what the bill still owes. A client naming an
	// amount is partially settling a bill, and may never name more than is outstanding.
	//
	// The CURRENCY is not here at all. It is the bill's, always, and a client-supplied one could
	// only ever agree with it or be wrong.
	Amount decimal.Decimal

	// Content is what the payer sees on their statement. Empty lets the provider fall back to the
	// order identifier, which is still traceable.
	Content string

	// IdempotencyKey lets a caller retry safely. Required in practice for a gateway collection: the
	// dangerous failure is a timeout AFTER the provider opened an order, where retrying without a key
	// shows the customer a second QR code for money they are already being asked for.
	IdempotencyKey string
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

	// AlreadyStarted says this collection was already open and nothing new was asked of the
	// provider. The instructions are the original ones, which is the point of retrying with a key.
	AlreadyStarted bool
}

// The refusal reasons opening a gateway collection can produce, beyond the ones RecordPayment
// already reports.
const (
	ReasonGatewayUnavailable    = "sales_payment.gateway_unavailable"
	ReasonMethodHasNoGateway    = "sales_payment.method_has_no_gateway"
	ReasonGatewayRefusedOrder   = "sales_payment.gateway_refused"
	ReasonGatewayOrderNotOpened = "sales_payment.gateway_order_not_opened"
	ReasonMethodNotAtPoint      = "sales_payment.method_not_accepted_at_point"
)

// StartGatewayPayment opens a collection with the provider and records the payment awaiting it.
func StartGatewayPayment(
	ctx corectx.Context,
	params StartGatewayPaymentParams,
	methods itExt.PaymentMethodExtService,
	orders itExt.PaymentOrderExtService,
	channelPayments itChannel.ChannelPaymentAppService,
	pointPayments *PointPaymentDomainServiceImpl,
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

	// The replay, checked BEFORE anything is validated or written. A retry of a collection that
	// already opened must answer with the same QR code: re-running the gates would be wasted work,
	// and reaching the provider again would open a second order for one debt.
	if params.IdempotencyKey != "" {
		existing, err := findPaymentByIdempotencyKey(ctx, params.SalesBillId, params.IdempotencyKey)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return replayGatewayPayment(existing), nil, nil
		}
	}

	// An order past its payment deadline takes no new payment: the goods it held are gone and the
	// customer may only cancel.
	if vErrs, err := assertOrderNotExpiredById(ctx, stringOf(bill, models.SalesBillFieldSalesOrderId)); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// The money is resolved from the BILL, never taken from the caller. A client that believes it
	// owes less than it does must not be able to make that true by saying so, and a currency it
	// names could only agree with the bill or be wrong.
	amount, vErrs := payableAmountOf(ctx, bill, params.Amount)
	if vErrs != nil {
		return nil, vErrs, nil
	}
	params.Amount = amount

	if vErrs, err := assertMethodCollectsThroughGateway(
		ctx, params, methods,
	); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	profileId, vErrs, err := paymentProfileOfBill(ctx, bill, params.PaymentMethodId, pointPayments)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// The payment is written first, and deliberately: it is the record that a collection was
	// attempted. Opening the order first and dying before the write would leave money collectable
	// against a bill with nothing awaiting it, which no sweep could then reconcile.
	recorded, vErrs, err := RecordPayment(ctx, RecordPaymentParams{
		SalesBillId:     params.SalesBillId,
		PaymentMethodId: params.PaymentMethodId,
		Amount:          params.Amount,
		CurrencyCode:    stringOf(bill, models.SalesBillFieldCurrencyCode),
		IdempotencyKey:  params.IdempotencyKey,
		Status:          string(models.SalesPaymentStatusPending),
	}, methods, channelPayments, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	opened, err := orders.CreatePayment(ctx, itExt.CreateGatewayPaymentCommand{
		OrgId:            stringOf(bill, models.SalesPaymentFieldOrgId),
		PaymentMethodId:  params.PaymentMethodId,
		PaymentProfileId: profileId,
		Amount:           params.Amount,
		Content:          params.Content,
		SalesPaymentId:   recorded.SalesPaymentId,
		SalesBillId:      params.SalesBillId,
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

// The refusal reasons resolving the amount can produce.
const (
	ReasonNothingOutstanding = "sales_bill.nothing_outstanding"
	ReasonAmountExceedsDue   = "sales_payment.amount_exceeds_outstanding"
)

// payableAmountOf answers what this collection is for: the caller's amount when it named one, and
// the whole outstanding balance when it did not.
//
// Overpayment is refused here even where the cash policy would allow it. A customer handing over a
// note and taking change is one thing; a QR code asking for more than the bill is owed is another,
// and the excess would have to be refunded through the provider rather than out of the till.
func payableAmountOf(
	ctx corectx.Context, bill dmodel.DynamicFields, requested decimal.Decimal,
) (decimal.Decimal, *ft.ClientErrors) {
	billId := stringOf(bill, models.SalesBillFieldId)
	captured, err := capturedTotalOf(ctx, billId)
	if err != nil {
		// Treated as nothing captured rather than failing the collection: the worst case is asking
		// for too much, which the outstanding check below then refuses on the caller's own number.
		captured = decimal.Zero
	}

	outstanding := decimalOf(bill, models.SalesBillFieldTotalAmount).Sub(captured)
	return resolvePayableAmount(outstanding, requested)
}

// resolvePayableAmount is the decision itself, kept separate from reading the money so it can be
// pinned by a test: the rule is what matters, and it must not depend on a database to be checkable.
func resolvePayableAmount(
	outstanding, requested decimal.Decimal,
) (decimal.Decimal, *ft.ClientErrors) {
	if !outstanding.IsPositive() {
		return decimal.Zero, refusal("sales_bill_id", ReasonNothingOutstanding,
			"this bill has nothing left to pay")
	}

	if !requested.IsPositive() {
		return outstanding, nil
	}
	if requested.GreaterThan(outstanding) {
		return decimal.Zero, refusal("amount", ReasonAmountExceedsDue,
			"this bill has "+outstanding.String()+" outstanding, less than the "+
				requested.String()+" requested")
	}
	return requested, nil
}

// replayGatewayPayment answers a retry with the collection that is already open, QR included.
//
// The instructions are read back from the payment rather than asked of the provider again: the
// customer may already be looking at that QR code, and a second one for the same debt is the
// confusion the idempotency key exists to prevent.
func replayGatewayPayment(existing dmodel.DynamicFields) *StartGatewayPaymentResult {
	return &StartGatewayPaymentResult{
		SalesPaymentId: stringOf(existing, models.SalesPaymentFieldId),
		SalesBillId:    stringOf(existing, models.SalesPaymentFieldSalesBillId),
		PaymentOrderId: stringOf(existing, models.SalesPaymentFieldPaymentOrderId),
		OrderCode:      stringOf(existing, models.SalesPaymentFieldProviderReference),
		QrCodeUrl:      stringOf(existing, models.SalesPaymentFieldQrCodeUrl),
		PayUrl:         stringOf(existing, models.SalesPaymentFieldPayUrl),
		AlreadyStarted: true,
	}
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

	// The instructions are stored, not just returned: a retry with the same key is answered from
	// this row, and a QR code that lived only in the lost response could not be handed back.
	if opened.QrCodeUrl != "" {
		fields[models.SalesPaymentFieldQrCodeUrl] = opened.QrCodeUrl
	}
	if opened.PayUrl != "" {
		fields[models.SalesPaymentFieldPayUrl] = opened.PayUrl
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

func paymentProfileOfBill(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	paymentMethodId string,
	pointPayments *PointPaymentDomainServiceImpl,
) (string, *ft.ClientErrors, error) {
	refuse := func(message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(
			"payment_method_id", ReasonMethodNotAtPoint, message))
		return vErrs
	}

	if pointPayments == nil {
		return "", refuse(
			"sales point payment mappings are unavailable, so no account can be resolved"), nil
	}

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId,
		stringOf(bill, models.SalesBillFieldSalesOrderId))
	if err != nil {
		return "", nil, err
	}
	if order == nil {
		return "", refuse(
			"this bill's sales order no longer exists, so its selling place cannot be resolved"), nil
	}

	salesPointId := stringOf(order, models.SalesOrderFieldSalesPointId)
	if salesPointId == "" {
		return "", refuse("this bill's sales order names no selling place"), nil
	}

	mapping, err := pointPayments.FindMapping(ctx, salesPointId, paymentMethodId)
	if err != nil {
		return "", nil, err
	}
	if mapping == nil {
		return "", refuse(
			"this payment method is not accepted at the selling place of this bill"), nil
	}

	return stringOf(mapping, models.SalesPointPaymentRelFieldPaymentProfileId), nil, nil
}
