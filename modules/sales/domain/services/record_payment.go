package services

import (
	"time"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Recording a payment (BR 39-42 and 75, CR 33/36/37/78, SALES-027).
//
// # Two gates on the method, not one (CR 33)
//
// A method must be BOTH mapped to the bill's channel AND currently usable. They are different
// questions with different owners: the mapping is Sales' own configuration, usability belongs to
// paymentinvoice and depends on the gateway registry of the running build as well as the row. A
// local mapping is not proof of usability - the deployment may simply not ship that adapter - and
// checking only the mapping would let a till accept a payment the gateway then refuses.
//
// # A replay returns SUCCESS (D-29)
//
// A duplicate external_transaction_id is not an error. A gateway told "conflict" retries forever,
// and each retry is another chance to record the money twice. The unique index is the mechanism; the
// contract is that the second call returns the payment the first one created.
//
// # CR 78: the method is revalidated on every NEW payment
//
// A bill that was open when a method was disabled must not become unpayable. Only the payment being
// added is checked, never the ones already recorded, so a partially-paid bill can legitimately be
// finished with a different method.

// RecordPaymentParams is what taking a payment needs.
type RecordPaymentParams struct {
	SalesBillId     string
	PaymentMethodId string

	Amount       decimal.Decimal
	CurrencyCode string

	// ExternalTransactionId is the provider's identifier. Empty for cash, which no provider issued
	// an id for - and which therefore has no replay protection, correctly, since a cash payment
	// cannot arrive twice by retry.
	ExternalTransactionId string
	ProviderReference     string

	// Status is what the provider says. Defaults to captured, because the common case at a till is
	// money already taken; a gateway flow supplies pending or authorized explicitly.
	Status string
}

// RecordPaymentResult is what recording produced.
type RecordPaymentResult struct {
	SalesPaymentId string
	SalesBillId    string

	CapturedTotal decimal.Decimal
	BillTotal     decimal.Decimal

	// ChangeDue is the excess when the cash-change policy permits overpayment (BR 42). It is NOT
	// part of the payment amount of the order: the customer hands over more and gets the difference
	// back, so counting it would overstate what the sale was worth.
	ChangeDue decimal.Decimal

	// AlreadyExisted marks the replay path.
	AlreadyExisted bool
}

// The refusal reasons recording a payment can produce.
const (
	ReasonMethodNotAllowedForChannel = "sales_payment.method_not_allowed_for_channel"
	ReasonMethodNotUsable            = "sales_payment.method_not_usable"
	ReasonTooManyMethods             = "sales_payment.too_many_methods"
	ReasonOverpaymentNotAllowed      = "sales_payment.overpayment_not_allowed"
	ReasonPaymentCurrencyMismatch    = "sales_payment.currency_mismatch"
	ReasonAmountNotPositive          = "sales_payment.amount_not_positive"
)

// RecordPayment takes money against a bill.
func RecordPayment(
	ctx corectx.Context,
	params RecordPaymentParams,
	methods itExt.PaymentMethodExtService,
	channelPayments itChannel.ChannelPaymentAppService,
	policy SalesPolicy,
) (*RecordPaymentResult, *ft.ClientErrors, error) {
	bill, err := loadRecord(ctx,
		models.SalesBillSchemaName, models.SalesBillFieldId, params.SalesBillId)
	if err != nil {
		return nil, nil, err
	}
	if bill == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("sales_bill_id", ReasonBillNotFound,
			"no bill exists with id '"+params.SalesBillId+"'"))
		return nil, vErrs, nil
	}

	// The replay check comes first, before any gate. A retry of a payment already recorded must
	// succeed even if the method has since been disabled - the money is already taken, and refusing
	// the acknowledgement would make the gateway retry forever against a bill that is already paid.
	if params.ExternalTransactionId != "" {
		existing, err := findPaymentByTransactionId(ctx,
			params.SalesBillId, params.ExternalTransactionId)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return replayResult(ctx, bill, existing)
		}
	}

	if vErrs, err := assertPaymentAcceptable(
		ctx, bill, params, methods, channelPayments, policy,
	); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	paymentId, err := writePayment(ctx, bill, params)
	if err != nil {
		if isUniqueViolation(err) && params.ExternalTransactionId != "" {
			// A concurrent retry won the race. The success path, not a failure.
			existing, lookupErr := findPaymentByTransactionId(ctx,
				params.SalesBillId, params.ExternalTransactionId)
			if lookupErr != nil {
				return nil, nil, lookupErr
			}
			if existing != nil {
				return replayResult(ctx, bill, existing)
			}
		}
		return nil, nil, err
	}

	captured, err := capturedTotalOf(ctx, params.SalesBillId)
	if err != nil {
		return nil, nil, err
	}
	billTotal := decimalOf(bill, models.SalesBillFieldTotalAmount)

	return &RecordPaymentResult{
		SalesPaymentId: paymentId,
		SalesBillId:    params.SalesBillId,
		CapturedTotal:  captured,
		BillTotal:      billTotal,
		ChangeDue:      changeDue(captured, billTotal, policy),
	}, nil, nil
}

// assertPaymentAcceptable runs every gate BR 39-42 and CR 33/36/37 put before a payment.
//
// Ordered cheapest-first: the local checks that need no other module run before the two that call
// across a boundary, so an obviously wrong request costs nothing.
func assertPaymentAcceptable(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	params RecordPaymentParams,
	methods itExt.PaymentMethodExtService,
	channelPayments itChannel.ChannelPaymentAppService,
	policy SalesPolicy,
) (*ft.ClientErrors, error) {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}

	if !models.NewSalesBillFrom(bill).IsOpen() {
		return refuse("sales_bill_id", ReasonBillNotOpen,
			"only an open bill may be paid; this one is '"+
				stringOf(bill, models.SalesBillFieldStatus)+"'"), nil
	}

	if !params.Amount.IsPositive() {
		return refuse("amount", ReasonAmountNotPositive,
			"a payment must be for more than zero"), nil
	}

	billCurrency := stringOf(bill, models.SalesBillFieldCurrencyCode)
	if params.CurrencyCode != "" && params.CurrencyCode != billCurrency {
		// Refused rather than converted: Sales has no FX, and a payment in another currency cannot
		// be reconciled against what the bill is owed.
		return refuse("currency_code", ReasonPaymentCurrencyMismatch,
			"this bill settles in "+billCurrency+", not "+params.CurrencyCode), nil
	}

	if vErrs, err := assertMethodCountWithinPolicy(ctx, params, policy); err != nil || vErrs != nil {
		return vErrs, err
	}
	if vErrs, err := assertOverpaymentPermitted(ctx, bill, params, policy); err != nil || vErrs != nil {
		return vErrs, err
	}

	// The two cross-module gates, last because they are the expensive ones.
	return assertMethodPermitted(ctx, bill, params, methods, channelPayments)
}

// assertMethodPermitted applies CR 33's two gates: mapped to the channel, AND usable.
func assertMethodPermitted(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	params RecordPaymentParams,
	methods itExt.PaymentMethodExtService,
	channelPayments itChannel.ChannelPaymentAppService,
) (*ft.ClientErrors, error) {
	refuse := func(reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("payment_method_id", reason, message))
		return vErrs
	}

	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId,
		stringOf(bill, models.SalesBillFieldSalesOrderId))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return refuse(ReasonMethodNotAllowedForChannel,
			"this bill's sales order no longer exists, so its channel cannot be checked"), nil
	}
	channelId := stringOf(order, models.SalesOrderFieldSalesChannelId)

	// Gate 1: the channel accepts this method. Default-deny (CR 76) — a channel with no mappings
	// accepts nothing, so a nil service refuses rather than waving everything through.
	if channelPayments == nil {
		return refuse(ReasonMethodNotAllowedForChannel,
			"payment method mappings are unavailable, so no method can be accepted"), nil
	}
	enabled, err := channelPayments.IsPaymentMethodEnabledForChannel(ctx,
		itChannel.IsPaymentMethodEnabledQuery{
			SalesChannelId:  channelId,
			PaymentMethodId: params.PaymentMethodId,
		})
	if err != nil {
		return nil, err
	}
	if enabled == nil || !enabled.HasData || !enabled.Data {
		return refuse(ReasonMethodNotAllowedForChannel,
			"this payment method is not enabled for the sales channel of this bill"), nil
	}

	// Gate 2: paymentinvoice can actually serve it. A mapping is not proof — usability depends on
	// the gateway registry of the running build and on the amount bounds, neither of which Sales
	// can see.
	if methods == nil {
		return refuse(ReasonMethodNotUsable,
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
		return refuse(ReasonMethodNotUsable,
			"this payment method cannot currently take a payment"), nil
	}
	if usable.ClientErrors.Count() > 0 {
		// The upstream reason travels through unchanged, so a till can tell "we stopped offering
		// that" from "that is over the limit for this method".
		out := ft.NewClientErrors()
		out.ConcatPtr(&usable.ClientErrors)
		return out, nil
	}
	return nil, nil
}

// assertMethodCountWithinPolicy enforces BR 39/40 and CR 37 together.
//
// The limit counts DISTINCT methods, not payments: three card taps on one card are one method, and a
// rule counting payments would refuse a legitimate split across two cards while allowing five taps
// on one.
func assertMethodCountWithinPolicy(
	ctx corectx.Context, params RecordPaymentParams, policy SalesPolicy,
) (*ft.ClientErrors, error) {
	if policy.MaxPaymentMethodsPerBill <= 0 {
		return nil, nil
	}

	existing, err := searchBy(ctx,
		models.SalesPaymentSchemaName, models.SalesPaymentFieldSalesBillId, params.SalesBillId)
	if err != nil {
		return nil, err
	}

	distinct := map[string]bool{params.PaymentMethodId: true}
	for _, payment := range existing {
		// A failed or cancelled payment does not occupy a slot: the customer's card was declined,
		// and refusing them a second method because of it would be perverse.
		status := stringOf(payment, models.SalesPaymentFieldStatus)
		if status == string(models.SalesPaymentStatusFailed) ||
			status == string(models.SalesPaymentStatusCancelled) {
			continue
		}
		distinct[stringOf(payment, models.SalesPaymentFieldPaymentMethodId)] = true
	}

	if int32(len(distinct)) <= policy.MaxPaymentMethodsPerBill {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("payment_method_id", ReasonTooManyMethods,
		"this bill already uses the maximum of "+
			decimal.NewFromInt(int64(policy.MaxPaymentMethodsPerBill)).String()+
			" payment methods"))
	return vErrs, nil
}

// assertOverpaymentPermitted enforces BR 42.
//
// Disallowed by default. When allow_cash_change is set the excess is accepted and change is computed
// - and the change is NOT counted as payment amount of the order, which is why ChangeDue is reported
// separately rather than folded into the captured total.
func assertOverpaymentPermitted(
	ctx corectx.Context,
	bill dmodel.DynamicFields,
	params RecordPaymentParams,
	policy SalesPolicy,
) (*ft.ClientErrors, error) {
	captured, err := capturedTotalOf(ctx, params.SalesBillId)
	if err != nil {
		return nil, err
	}

	would := captured.Add(params.Amount)
	billTotal := decimalOf(bill, models.SalesBillFieldTotalAmount)
	if !would.GreaterThan(billTotal) {
		return nil, nil
	}
	if policy.AllowOverpayment || policy.AllowCashChange {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("amount", ReasonOverpaymentNotAllowed,
		"this payment would take "+would.String()+" against a bill of "+billTotal.String()+
			"; overpayment is not permitted"))
	return vErrs, nil
}

// changeDue is the excess handed back, or zero.
func changeDue(captured, billTotal decimal.Decimal, policy SalesPolicy) decimal.Decimal {
	if !policy.AllowCashChange {
		return decimal.Zero
	}
	if excess := captured.Sub(billTotal); excess.IsPositive() {
		return excess
	}
	return decimal.Zero
}

// capturedTotalOf sums the money actually in against a bill.
func capturedTotalOf(ctx corectx.Context, billId string) (decimal.Decimal, error) {
	payments, err := searchBy(ctx,
		models.SalesPaymentSchemaName, models.SalesPaymentFieldSalesBillId, billId)
	if err != nil {
		return decimal.Zero, err
	}
	return models.SumCapturedAmount(payments), nil
}

// findPaymentByTransactionId looks for a payment already recorded under a provider's id.
func findPaymentByTransactionId(
	ctx corectx.Context, billId, transactionId string,
) (dmodel.DynamicFields, error) {
	payments, err := searchBy(ctx,
		models.SalesPaymentSchemaName, models.SalesPaymentFieldSalesBillId, billId)
	if err != nil {
		return nil, err
	}
	for _, payment := range payments {
		if stringOf(payment, models.SalesPaymentFieldExternalTransactionId) == transactionId {
			return payment, nil
		}
	}
	return nil, nil
}

// replayResult answers a duplicate with the payment that already exists.
func replayResult(
	ctx corectx.Context, bill dmodel.DynamicFields, existing dmodel.DynamicFields,
) (*RecordPaymentResult, *ft.ClientErrors, error) {
	billId := stringOf(bill, models.SalesBillFieldId)
	captured, err := capturedTotalOf(ctx, billId)
	if err != nil {
		return nil, nil, err
	}
	return &RecordPaymentResult{
		SalesPaymentId: stringOf(existing, models.SalesPaymentFieldId),
		SalesBillId:    billId,
		CapturedTotal:  captured,
		BillTotal:      decimalOf(bill, models.SalesBillFieldTotalAmount),
		AlreadyExisted: true,
	}, nil, nil
}

// writePayment inserts the payment row.
func writePayment(
	ctx corectx.Context, bill dmodel.DynamicFields, params RecordPaymentParams,
) (string, error) {
	engine, err := engineFor(models.SalesPaymentSchemaName)
	if err != nil {
		return "", err
	}
	id, err := model.NewId()
	if err != nil {
		return "", err
	}

	status := params.Status
	if status == "" {
		status = string(models.SalesPaymentStatusCaptured)
	}
	currency := params.CurrencyCode
	if currency == "" {
		currency = stringOf(bill, models.SalesBillFieldCurrencyCode)
	}

	fields := dmodel.DynamicFields{
		models.SalesPaymentFieldId:              string(*id),
		models.SalesPaymentFieldSalesBillId:     params.SalesBillId,
		models.SalesPaymentFieldPaymentMethodId: params.PaymentMethodId,
		models.SalesPaymentFieldAmount:          params.Amount,
		models.SalesPaymentFieldCurrencyCode:    currency,
		models.SalesPaymentFieldStatus:          status,
		basemodel.FieldOrgId:                    stringOf(bill, basemodel.FieldOrgId),
	}
	if params.ExternalTransactionId != "" {
		fields[models.SalesPaymentFieldExternalTransactionId] = params.ExternalTransactionId
	}
	if params.ProviderReference != "" {
		fields[models.SalesPaymentFieldProviderReference] = params.ProviderReference
	}
	if status == string(models.SalesPaymentStatusCaptured) {
		fields[models.SalesPaymentFieldPaidAt] = model.ModelDateTime(time.Now().UTC())
	}

	if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
		return "", err
	}

	// ONLY a captured payment is announced. An authorization is a hold the provider may still
	// release, and telling consumers money arrived when it has not is the same mistake as counting
	// it toward settlement - which SumCapturedAmount already refuses to do.
	if status == string(models.SalesPaymentStatusCaptured) {
		if _, err := RecordEvent(ctx, RecordEventParams{
			EventType:   models.EventSalesPaymentCaptured,
			AggregateId: params.SalesBillId,
			OrgId:       stringOf(bill, basemodel.FieldOrgId),
			Payload: map[string]any{
				"sales_payment_id":  string(*id),
				"sales_bill_id":     params.SalesBillId,
				"sales_order_id":    stringOf(bill, models.SalesBillFieldSalesOrderId),
				"payment_method_id": params.PaymentMethodId,
				"amount":            params.Amount,
				"currency_code":     currency,
			},
		}); err != nil {
			return "", err
		}
	}
	return string(*id), nil
}
