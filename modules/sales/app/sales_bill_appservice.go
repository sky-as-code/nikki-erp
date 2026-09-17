package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itBilling "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/billing"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The bill's authorized surface. As with the order, the bodies below are the former engine action
// callbacks moved verbatim, and the ports they need are fields rather than package globals.
//
// The distributed lock is passed to the operation services untouched. Split and merge acquire it
// themselves, before opening their transaction and releasing it after the commit -- a lock that
// did not outlive its transaction would let a second holder read pre-commit state, which is the
// race it exists to prevent.

type SalesBillApplicationServiceImpl struct {
	composable.CrudApplicationService

	effectiveSettings itExt.EffectiveSettingsExtService
	orderLock         distributedlock.DistributedLock
	paymentMethods    itExt.PaymentMethodExtService
	paymentOrders     itExt.PaymentOrderExtService

	// The channel mapping gate answers "does this channel accept this method"; paymentMethods
	// answers "is the method usable at all". Both are consulted before money moves.
	channelPayments itChannel.ChannelPaymentAppService

	pointPayments *services.PointPaymentDomainServiceImpl
}

func NewSalesBillApplicationService(
	base composable.CrudApplicationService,
	settings itExt.EffectiveSettingsExtService,
	dLock distributedlock.DistributedLock,
	methods itExt.PaymentMethodExtService,
	orders itExt.PaymentOrderExtService,
	channels itChannel.ChannelPaymentAppService,
	pointPayments *services.PointPaymentDomainServiceImpl,
) itBilling.SalesBillApplicationService {
	return &SalesBillApplicationServiceImpl{
		CrudApplicationService: base,
		effectiveSettings:      settings,
		orderLock:              dLock,
		paymentMethods:         methods,
		paymentOrders:          orders,
		channelPayments:        channels,
		pointPayments:          pointPayments,
	}
}

func (this *SalesBillApplicationServiceImpl) Split(
	ctx corectx.Context, cmd itBilling.BillActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSplitBill, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runSplitBill(ctx, cmd)
}

// Merge is collection-level: it names several source bills rather than addressing one, so there is
// no single record to place in an org.
func (this *SalesBillApplicationServiceImpl) Merge(
	ctx corectx.Context, cmd itBilling.BillActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, PermissionMergeBill, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runMergeBills(ctx, cmd)
}

func (this *SalesBillApplicationServiceImpl) Pay(
	ctx corectx.Context, cmd itBilling.BillActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionPayBill, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runPayBill(ctx, cmd)
}

func (this *SalesBillApplicationServiceImpl) Settle(
	ctx corectx.Context, cmd itBilling.BillActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionSettleBill, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runSettleBill(ctx, cmd)
}

// Starting a gateway payment shares the pay permission: handing the customer to a provider is the
// same power over the same money as recording a payment directly.
func (this *SalesBillApplicationServiceImpl) StartGatewayPayment(
	ctx corectx.Context, cmd itBilling.BillActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionPayBill, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runStartGatewayPayment(ctx, cmd)
}

// The read-only billing resources.

func NewSalesBillLineApplicationService(base composable.CrudApplicationService) itBilling.SalesBillLineApplicationService {
	return &SalesBillLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesBillLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesBillRelationApplicationService(base composable.CrudApplicationService) itBilling.SalesBillRelationApplicationService {
	return &SalesBillRelationApplicationServiceImpl{CrudApplicationService: base}
}

type SalesBillRelationApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPaymentApplicationService(base composable.CrudApplicationService) itBilling.SalesPaymentApplicationService {
	return &SalesPaymentApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPaymentApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesFulfillmentRequestApplicationService(base composable.CrudApplicationService) itBilling.SalesFulfillmentRequestApplicationService {
	return &SalesFulfillmentRequestApplicationServiceImpl{CrudApplicationService: base}
}

type SalesFulfillmentRequestApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesFulfillmentRequestLineApplicationService(base composable.CrudApplicationService) itBilling.SalesFulfillmentRequestLineApplicationService {
	return &SalesFulfillmentRequestLineApplicationServiceImpl{CrudApplicationService: base}
}

type SalesFulfillmentRequestLineApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// The moved action bodies follow.

func (this *SalesBillApplicationServiceImpl) runSplitBill(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	result, vErrs, err := services.SplitBill(ctx, services.SplitBillParams{
		SourceBillId: readStringParam(params, paramRecordId),
		Parts:        readSplitParts(params),
	}, this.orderLock, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"source_bill_id":   result.SourceBillId,
			"created_bill_ids": result.CreatedBillIds,

			// Both totals, so a caller can see the split preserved the sum.
			"total_before": result.TotalBefore,
			"total_after":  result.TotalAfter,
		},
	}, nil
}

func (this *SalesBillApplicationServiceImpl) runMergeBills(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.MergeBills(ctx, services.MergeBillParams{
		SourceBillIds: readStringsParam(params, "source_bill_ids"),
	}, this.orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"merged_bill_id":  result.MergedBillId,
			"source_bill_ids": result.SourceBillIds,
			"total_before":    result.TotalBefore,
			"total_after":     result.TotalAfter,
		},
	}, nil
}

func readSplitParts(params map[string]any) []services.SplitBillPart {
	raw, ok := params["parts"].([]any)
	if !ok {
		return nil
	}

	parts := make([]services.SplitBillPart, 0, len(raw))
	for _, item := range raw {
		fields, ok := item.(map[string]any)
		if !ok {
			continue
		}
		allocations := map[string]decimal.Decimal{}
		if raw, ok := fields["allocations"].(map[string]any); ok {
			for lineId := range raw {
				allocations[lineId] = readDecimalParam(raw, lineId)
			}
		}
		parts = append(parts, services.SplitBillPart{Allocations: allocations})
	}
	return parts
}

// runPayBill records a payment and settles immediately rather than in a second request: a bill
// whose last payment just landed is settled by that fact.
func (this *SalesBillApplicationServiceImpl) runPayBill(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)
	billId := readStringParam(params, paramRecordId)

	result, vErrs, err := services.RecordPayment(ctx, services.RecordPaymentParams{
		SalesBillId:           billId,
		PaymentMethodId:       readStringParam(params, "payment_method_id"),
		Amount:                readDecimalParam(params, "amount"),
		CurrencyCode:          readStringParam(params, "currency_code"),
		ExternalTransactionId: readStringParam(params, "external_transaction_id"),
		ProviderReference:     readStringParam(params, "provider_reference"),
		Status:                readStringParam(params, "status"),
	}, this.paymentMethods, this.channelPayments, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	settled, _, err := services.SettleBillIfPaid(ctx, billId)
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"sales_payment_id": result.SalesPaymentId,
		"sales_bill_id":    result.SalesBillId,
		"captured_total":   result.CapturedTotal,
		"bill_total":       result.BillTotal,

		// Never folded into the captured total: change is handed back, so counting it would
		// overstate what the sale was worth.
		"change_due":      result.ChangeDue,
		"already_existed": result.AlreadyExisted,
	}
	if settled != nil {
		data["bill_status"] = settled.Status
		data["payment_status"] = settled.PaymentStatus
	}
	return &dyn.OpResult[any]{HasData: true, Data: data}, nil
}

// runSettleBill is exposed separately from pay so an operator can reconcile a bill whose
// payments arrived by a path Sales did not record.
func (this *SalesBillApplicationServiceImpl) runSettleBill(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	billId := readStringParam(params, paramRecordId)

	result, vErrs, err := services.SettleBillIfPaid(ctx, billId)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_bill_id":  result.SalesBillId,
			"status":         result.Status,
			"payment_status": result.PaymentStatus,
			"captured_total": result.CapturedTotal,
			"bill_total":     result.BillTotal,
			"settled":        result.Settled,
		},
	}, nil
}

func (this *SalesBillApplicationServiceImpl) runStartGatewayPayment(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	// No currency_code is read from the body, deliberately: it is the bill's, and a caller naming one
	// could only ever agree with it or be wrong. amount stays optional - omitted means the whole
	// outstanding balance.
	result, vErrs, err := services.StartGatewayPayment(ctx, services.StartGatewayPaymentParams{
		SalesBillId:     readStringParam(params, paramRecordId),
		PaymentMethodId: readStringParam(params, paramPaymentMethodId),
		Amount:          readDecimalParam(params, "amount"),
		Content:         readStringParam(params, "content"),
		IdempotencyKey:  readStringParam(params, "idempotency_key"),
	}, this.paymentMethods, this.paymentOrders, this.channelPayments, this.pointPayments, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	// No bill status here, unlike pay: nothing has been settled. The customer has been handed
	// something to pay with, and the bill moves when the provider says the money arrived.
	return &dyn.OpResult[any]{HasData: true, Data: map[string]any{
		"sales_payment_id": result.SalesPaymentId,
		"sales_bill_id":    result.SalesBillId,
		"payment_order_id": result.PaymentOrderId,
		"order_code":       result.OrderCode,
		"qr_code_url":      result.QrCodeUrl,
		"pay_url":          result.PayUrl,

		// True means this retry was recognised and the instructions are the original ones, so a
		// caller can tell "your QR is still valid" from "here is a new one".
		"already_started": result.AlreadyStarted,
	}}, nil
}
