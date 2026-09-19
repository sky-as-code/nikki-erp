package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itOrder "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/order"
	"time"
)

// The sales order's authorized surface.
//
// Every method here asserts its permission and then delegates: the bodies below are the former
// engine action callbacks, moved verbatim. The ports they need are struct fields rather than
// package globals, which is the whole point of the move -- a legacy action callback was handed no
// dependencies, so the ports had to be pushed into the package before any request arrived.

type SalesOrderCrudApplicationServiceImpl struct {
	composable.CrudApplicationService
	orderSvc itOrder.SalesOrderDomainService

	taxCalculation    itExt.TaxCalculationExtService
	effectiveSettings itExt.EffectiveSettingsExtService
	orderLock         distributedlock.DistributedLock
	productVariants   itExt.ProductVariantExtService
	orderFulfillment  itExt.FulfillmentExtService
	pricingBasis      itExt.ProductPricingBasisExtService

	// Nil is the ordinary case for a deployment that sells nothing from a machine; the confirm path
	// then behaves exactly as it did before fulfillment existed.
	fulfillmentReservations itExt.FulfillmentReservationExtService

	partyPort itExt.PartyExtService

	// The refund ports, for the cancel of a paid order that raises and may dispatch a refund.
	paymentOrders itExt.PaymentOrderExtService
	invoicing     itInvoicing.InvoicingExtService
}

// SalesOrderApplicationServiceImpl is the receiver name the moved bodies use.
type SalesOrderApplicationServiceImpl = SalesOrderCrudApplicationServiceImpl

func NewSalesOrderCrudApplicationService(
	base composable.CrudApplicationService,
	tax itExt.TaxCalculationExtService,
	settings itExt.EffectiveSettingsExtService,
	dLock distributedlock.DistributedLock,
	products itExt.ProductVariantExtService,
	fulfillment itExt.FulfillmentExtService,
	basis itExt.ProductPricingBasisExtService,
	reservations itExt.FulfillmentReservationExtService,
	parties itExt.PartyExtService,
	paymentOrders itExt.PaymentOrderExtService,
	invoicing itInvoicing.InvoicingExtService,
) itOrder.SalesOrderApplicationService {
	return &SalesOrderCrudApplicationServiceImpl{
		CrudApplicationService:  base,
		orderSvc:                base.DomainService().(itOrder.SalesOrderDomainService),
		taxCalculation:          tax,
		effectiveSettings:       settings,
		orderLock:               dLock,
		productVariants:         products,
		orderFulfillment:        fulfillment,
		pricingBasis:            basis,
		fulfillmentReservations: reservations,
		partyPort:               parties,
		paymentOrders:           paymentOrders,
		invoicing:               invoicing,
	}
}

// The authorized entry points. Each asserts, then runs the moved body.
//
// CreateOrder is collection-level: there is no record to place in an org yet, so it asserts the
// action without the record check the others make.

func (this *SalesOrderCrudApplicationServiceImpl) CreateOrder(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, PermissionCreateOrder, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runCreateOrder(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) Reprice(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionReprice, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runReprice(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) Confirm(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionConfirm, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runConfirmOrder(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) Cancel(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionCancel, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCancelOrder(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) ApplyVoucher(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionApplyVoucher, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runApplyVoucher(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) ExplainPrice(
	ctx corectx.Context, query itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionExplainPrice, query); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runExplainPrice(ctx, query)
}

func (this *SalesOrderCrudApplicationServiceImpl) GrantManualDiscount(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionManualDiscount, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runGrantManualDiscount(ctx, cmd)
}

// Revoking shares the grant permission: whoever may move the price may move it back.
func (this *SalesOrderCrudApplicationServiceImpl) RevokeManualDiscount(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionManualDiscount, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runRevokeManualDiscount(ctx, cmd)
}

// The moved action bodies follow, unchanged except for the receiver, the result type and the ports
// now being fields.

// runExplainPrice answers why an order costs what it costs. A POST only because that is what
// ActionTypeGeneric routes are here; it writes nothing and is safely repeatable.
func (this *SalesOrderApplicationServiceImpl) runExplainPrice(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	orderId := readStringParam(params, paramRecordId)

	explanation, err := services.ExplainOrderPrice(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if explanation == nil {
		return &dyn.OpResult[any]{
			ClientErrors: *services.OrderNotFoundErrors(orderId),
		}, nil
	}

	lines := make([]map[string]any, 0, len(explanation.Lines))
	for _, line := range explanation.Lines {
		lines = append(lines, map[string]any{
			"sales_order_line_id": line.SalesOrderLineId,
			"line_number":         line.LineNumber,
			"product_code":        line.ProductCode,
			"product_name":        line.ProductName,
			"quantity":            line.Quantity,
			"base_amount":         line.BaseAmount,
			"steps":               stepsPayload(line.Steps),
			"net_amount":          line.NetAmount,
			"tax_amount":          line.TaxAmount,
			"final_amount":        line.FinalAmount,

			// False means the stored adjustments do not account for the stored net — a bug worth
			// surfacing rather than displaying an explanation that does not add up.
			"steps_reconcile": line.StepsReconcile(),
		})
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_order_id": explanation.SalesOrderId,
			"lines":          lines,
			"order_steps":    stepsPayload(explanation.OrderSteps),
			"subtotal":       explanation.Subtotal,
			"discount_total": explanation.DiscountTotal,
			"tax_total":      explanation.TaxTotal,
			"grand_total":    explanation.GrandTotal,
		},
	}, nil
}

func stepsPayload(steps []services.PriceStep) []map[string]any {
	payload := make([]map[string]any, 0, len(steps))
	for _, step := range steps {
		payload = append(payload, map[string]any{
			"sequence":    step.Sequence,
			"type":        step.Type,
			"source_type": step.SourceType,
			"source_id":   step.SourceId,
			"description": step.Description,
			"base_amount": step.BaseAmount,
			"amount":      step.Amount,
		})
	}
	return payload
}

// runApplyVoucher is transport only; every rule lives in services.ApplyVoucher so the operation
// stays reachable from CQRS and from another module's port without HTTP.
func (this *SalesOrderApplicationServiceImpl) runApplyVoucher(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	orderId := readStringParam(params, paramRecordId)

	order, err := services.LoadSalesOrderForVoucher(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		// A client error rather than a fault: the caller named a record that does not exist.
		return &dyn.OpResult[any]{
			ClientErrors: *services.OrderNotFoundErrors(orderId),
		}, nil
	}

	// The clock is read once at the edge and passed inward, so a test can reproduce it and two gates
	// in the same request cannot disagree about the time.
	nowUnix := readInt64Param(params, paramNowUnix)
	if nowUnix == 0 {
		nowUnix = services.NowUnix()
	}

	result, vErrs, err := services.ApplyVoucher(ctx, services.ApplyVoucherParams{
		Code:              readStringParam(params, paramVoucherCode),
		SalesOrderId:      orderId,
		OrgId:             order.OrgId,
		SalesChannelId:    order.SalesChannelId,
		SalesPointId:      order.SalesPointId,
		AppliedProgramIds: order.AppliedProgramIds,

		// An empty basket still evaluates: a program with no conditions applies, one with a minimum
		// spend does not.
		Facts: pricing.BasketFacts{
			Subtotal:      order.Subtotal,
			TotalQuantity: order.TotalQuantity,
			NowUnix:       nowUnix,
		},
		NowUnix: nowUnix,
	})
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"program_id":            result.ProgramId,
			"accepted_program_ids":  result.AcceptedProgramIds,
			"displaced_program_ids": result.DisplacedProgramIds,
		},
	}, nil
}

// readInt64Param accepts every numeric shape because a value that arrived as JSON is a float64, and
// a reader taking only int64 would silently ignore it.
func readInt64Param(params map[string]any, field string) int64 {
	value, ok := params[field]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int64:
		return typed
	case int32:
		return int64(typed)
	case int:
		return int64(typed)
	case float64:
		return int64(typed)
	}
	return 0
}

// runReprice is a separate action rather than a hook in the line CRUD routes, so a caller adding
// three lines gets one reprice rather than three adjustment chains nobody saw. A caller can forget
// it, leaving stale draft totals, but confirm reprices unconditionally before freezing anything.
func (this *SalesOrderApplicationServiceImpl) runReprice(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	orderId := readStringParam(params, paramRecordId)

	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	result, vErrs, err := services.RepriceOrder(ctx, orderId, this.taxCalculation, policy, this.pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"subtotal":       result.Subtotal,
			"discount_total": result.DiscountTotal,
			"tax_total":      result.TaxTotal,
			"grand_total":    result.GrandTotal,
			"line_count":     result.LineCount,
		},
	}, nil
}

func (this *SalesOrderApplicationServiceImpl) runCreateOrder(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	request := services.CreateOrderParams{
		SalesChannelCode:  readStringParam(params, "sales_channel_code"),
		SalesPointId:      readStringParam(params, "sales_point_id"),
		CustomerReference: readStringParam(params, "customer_reference"),
		CurrencyCode:      readStringParam(params, "currency_code"),
		ExternalReference: readStringParam(params, "external_reference"),
		IdempotencyKey:    readStringParam(params, "idempotency_key"),
		Lines:             readOrderLines(params),

		EstimatedTotalPrice: readOptionalDecimalParam(params, "estimated_total_price"),
	}
	validUntil, vErrs := readOptionalDateTimeParam(params, "valid_until")
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	request.ValidUntil = validUntil

	result, vErrs, err := services.CreateOrder(ctx, request, this.taxCalculation, this.productVariants, this.pricingBasis, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	data := map[string]any{
		"sales_order_id":   result.SalesOrderId,
		"order_number":     result.OrderNumber,
		"sales_channel_id": result.SalesChannelId,

		// True on the idempotent replay path. The caller gets a success either way; this says
		// whether anything was actually written.
		"already_existed": result.AlreadyExisted,
	}
	if result.Pricing != nil {
		data["subtotal"] = result.Pricing.Subtotal
		data["discount_total"] = result.Pricing.DiscountTotal
		data["tax_total"] = result.Pricing.TaxTotal
		data["grand_total"] = result.Pricing.GrandTotal
	}

	// The draft is saved and announced; only now, and only when its snapshot says so, is it
	// confirmed, through the same action a user would call. A refused auto-confirm leaves the
	// draft and reports why beside it: the create succeeded, the confirm did not.
	if result.AutoConfirmOrder && !result.AlreadyExisted {
		data["auto_confirm"] = this.autoConfirm(ctx, result.SalesOrderId, policy)
	}

	return &dyn.OpResult[any]{HasData: true, Data: data}, nil
}

// autoConfirm confirms a freshly created draft on the channel's behalf and reports the outcome.
func (this *SalesOrderApplicationServiceImpl) autoConfirm(
	ctx corectx.Context, orderId string, policy services.SalesPolicy,
) map[string]any {
	confirmed, vErrs, err := services.ConfirmOrderWith(ctx, orderId,
		services.ConfirmOrderOptions{Automatic: true},
		this.orderLock, this.taxCalculation, this.orderFulfillment, this.pricingBasis,
		policy, services.FulfillmentMethodService(), this.fulfillmentReservations)
	if err != nil {
		return map[string]any{"attempted": true, "confirmed": false, "error": err.Error()}
	}
	if vErrs != nil {
		return map[string]any{"attempted": true, "confirmed": false, "client_errors": *vErrs}
	}
	outcome := map[string]any{
		"attempted":       true,
		"confirmed":       true,
		"status":          confirmed.Status,
		"initial_bill_id": confirmed.InitialBillId,
		"pending":         confirmed.Pending,
	}
	if view, err := services.LoadConfirmedOrderView(ctx, orderId, confirmed.InitialBillId); err == nil && view != nil {
		outcome["order"] = view.Order
		outcome["initial_bill"] = view.Bill
	}
	return outcome
}

// readOrderLines skips a line whose shape is not a map rather than erroring; the validation that
// follows refuses an empty basket with a business message instead of a parser complaint.
func readOrderLines(params map[string]any) []services.CreateOrderLine {
	raw, ok := params["lines"].([]any)
	if !ok {
		return nil
	}

	lines := make([]services.CreateOrderLine, 0, len(raw))
	for _, item := range raw {
		fields, ok := item.(map[string]any)
		if !ok {
			continue
		}
		lines = append(lines, services.CreateOrderLine{
			ProductVariantId: readStringParam(fields, "product_variant_id"),
			UomId:            readStringParam(fields, "uom_id"),
			Quantity:         readDecimalParam(fields, "quantity"),
			UnitPrice:        readDecimalParam(fields, "unit_price"),
			ProductCode:      readStringParam(fields, "product_code"),
			ProductName:      readStringParam(fields, "product_name"),
			EstimatedPrice:   readOptionalDecimalParam(fields, "estimated_price"),
		})
	}
	return lines
}

func (this *SalesOrderApplicationServiceImpl) runConfirmOrder(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	result, vErrs, err := services.ConfirmOrderWith(ctx,
		readStringParam(params, paramRecordId),
		services.ConfirmOrderOptions{ConfirmationNote: readStringParam(params, "confirmation_note")},
		this.orderLock, this.taxCalculation, this.orderFulfillment, this.pricingBasis,
		policy, services.FulfillmentMethodService(), this.fulfillmentReservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	data := map[string]any{
		"sales_order_id":       result.SalesOrderId,
		"status":               result.Status,
		"confirmed_at":         result.ConfirmedAt,
		"initial_bill_id":      result.InitialBillId,
		"already_confirmed":    result.AlreadyConfirmed,
		"redeemed_voucher_ids": result.RedeemedVoucherIds,

		// In the response, not just the log: a kiosk that believed a confirm was complete would
		// dispense goods against an order with no fulfilment request.
		"pending": result.Pending,
	}
	if result.Pricing != nil {
		data["grand_total"] = result.Pricing.GrandTotal
	}

	// The whole order and the whole bill, not just their ids: the caller's next act is to ask for
	// money, and a second round trip to learn how much would be one the confirm could have saved.
	view, err := services.LoadConfirmedOrderView(ctx, result.SalesOrderId, result.InitialBillId)
	if err != nil {
		return nil, err
	}
	if view != nil {
		data["order"] = view.Order
		data["initial_bill"] = view.Bill
	}

	return &dyn.OpResult[any]{HasData: true, Data: data}, nil
}

func (this *SalesOrderApplicationServiceImpl) runCancelOrder(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.CancelOrderWith(ctx,
		readStringParam(params, paramRecordId),
		services.CancelOrderOptions{
			Reason:           readStringParam(params, "reason"),
			CancellationNote: readStringParam(params, "cancellation_note"),
			Policy:           services.ResolveSalesPolicy(ctx, this.effectiveSettings),
			RefundDeps:       this.refundDeps(),
		},
		this.orderLock, this.fulfillmentReservations)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_order_id":           result.SalesOrderId,
			"status":                   result.Status,
			"cancelled_at":             result.CancelledAt,
			"released_voucher_ids":     result.ReleasedVoucherIds,
			"released_fulfillment_ids": result.ReleasedFulfillmentIds,
			"pending":                  result.Pending,
			"refund_request_id":        result.RefundRequestId,
			"refund_status":            result.RefundStatus,
		},
	}, nil
}

// refundDeps bundles the ports a cancel needs to raise and dispatch a paid order's refund.
func (this *SalesOrderApplicationServiceImpl) refundDeps() services.RefundProcessingDeps {
	return services.RefundProcessingDeps{
		Fulfillment:   this.orderFulfillment,
		Invoicing:     this.invoicing,
		PaymentOrders: this.paymentOrders,
	}
}

// processGrantManualDiscount records an operator override and reprices.
func (this *SalesOrderApplicationServiceImpl) runGrantManualDiscount(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	result, vErrs, err := services.GrantManualDiscount(ctx, services.GrantManualDiscountParams{
		SalesOrderId:     readStringParam(params, paramRecordId),
		SalesOrderLineId: readStringParam(params, "sales_order_line_id"),
		Amount:           readDecimalParam(params, "discount_amount"),
		Reason:           readStringParam(params, "reason"),
	}, this.taxCalculation, policy, this.pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"sales_manual_discount_id": result.SalesManualDiscountId,
			"sales_order_id":           result.SalesOrderId,

			// Both totals, because the engine caps a discount at what is owed: the difference is not
			// always the amount asked for.
			"total_before": result.TotalBefore,
			"total_after":  result.TotalAfter,
		},
	}, nil
}

// processRevokeManualDiscount withdraws an override and reprices.
func (this *SalesOrderApplicationServiceImpl) runRevokeManualDiscount(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	policy := services.ResolveSalesPolicy(ctx, this.effectiveSettings)

	vErrs, err := services.RevokeManualDiscount(ctx,
		readStringParam(params, paramRecordId),
		readStringParam(params, "sales_manual_discount_id"),
		this.taxCalculation, policy, this.pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{
		HasData: true,
		Data:    map[string]any{"sales_order_id": readStringParam(params, paramRecordId)},
	}, nil
}

// The party assignments. Four permissions rather than one: who is billed, who pays and who bought
// are answered by different people in a credit-managed sale, so a grantor withholds them
// separately.

func (this *SalesOrderCrudApplicationServiceImpl) AssignParties(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionAssignParties, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(cmd, paramRecordId),
		SoldTo:       readPartyAssignment(cmd, paramSoldToPartyId),
		BillTo:       readPartyAssignment(cmd, paramBillToPartyId),
		Payer:        readPartyAssignment(cmd, paramPayerPartyId),
	})
}

func (this *SalesOrderCrudApplicationServiceImpl) AssignSoldToParty(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionAssignSoldTo, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(cmd, paramRecordId),
		SoldTo:       readPartyAssignment(cmd, paramSoldToPartyId),
	})
}

func (this *SalesOrderCrudApplicationServiceImpl) AssignBillToParty(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionAssignBillTo, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(cmd, paramRecordId),
		BillTo:       readPartyAssignment(cmd, paramBillToPartyId),
	})
}

func (this *SalesOrderCrudApplicationServiceImpl) AssignPayerParty(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionAssignPayer, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runPartyAssignment(ctx, services.AssignPartiesParams{
		SalesOrderId: readStringParam(cmd, paramRecordId),
		Payer:        readPartyAssignment(cmd, paramPayerPartyId),
	})
}

func (this *SalesOrderCrudApplicationServiceImpl) runPartyAssignment(
	ctx corectx.Context, params services.AssignPartiesParams,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.AssignParties(ctx, params, this.partyPort)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{HasData: false}, nil
}

// readPartyAssignment distinguishes "not mentioned" from "explicitly cleared": a field present but
// null asks to unset the party, which is not the same as leaving it alone.
func readPartyAssignment(params dmodel.DynamicFields, field string) services.PartyAssignment {
	value, present := params[field]
	if !present {
		return services.PartyAssignment{}
	}
	if value == nil {
		return services.PartyAssignment{Requested: true}
	}
	return services.PartyAssignment{
		Requested: true,
		PartyId:   readStringParam(params, field),
	}
}

// The order's two fulfillment views. Both are reads: one lists what a given order owes, the other
// asks which outlets could satisfy a basket. The search is collection-level -- it names no order,
// so there is no record to place in an org.

func (this *SalesOrderCrudApplicationServiceImpl) ListFulfillments(
	ctx corectx.Context, query itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionRead, query); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runOrderFulfillments(ctx, query)
}

func (this *SalesOrderCrudApplicationServiceImpl) SearchFulfillmentTargets(
	ctx corectx.Context, query itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, composable.PermissionRead, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runFulfillmentTargetSearch(ctx, query)
}

// processOrderFulfillments answers what an order's deliveries owe. Refund state is deliberately
// absent from fulfillment_status; the quantities carry it instead, and pending_refund_qty is what
// separates "still owed" from "may be attempted now".
func (this *SalesOrderCrudApplicationServiceImpl) runOrderFulfillments(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	views, err := services.ViewOrderFulfillments(ctx, readStringParam(params, paramRecordId))
	if err != nil {
		return nil, err
	}

	fulfillments := make([]map[string]any, 0, len(views))
	for _, view := range views {
		items := make([]map[string]any, 0, len(view.Items))
		for _, item := range view.Items {
			items = append(items, map[string]any{
				"fulfillment_item_id": item.ItemId,
				"sales_order_line_id": item.SalesOrderLineId,
				"product_variant_id":  item.ProductVariantId,
				"item_status":         item.Status,
				"ordered_qty":         item.OrderedQty,
				"fulfilled_qty":       item.FulfilledQty,
				"refunded_qty":        item.RefundedQty,
				"remaining_qty":       item.RemainingQty,
				"pending_refund_qty":  item.PendingRefundQty,
				"fulfillable_qty":     item.FulfillableQty,
			})
		}
		fulfillments = append(fulfillments, map[string]any{
			"fulfillment_id":         view.FulfillmentId,
			"fulfillment_method_id":  view.MethodId,
			"fulfillment_type":       view.Type,
			"target_outlet_id":       view.TargetOutletId,
			"fulfillment_status":     view.Status,
			"reservation_expires_at": view.ReservationExpiresAt,
			"items":                  items,
		})
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data:    map[string]any{"fulfillments": fulfillments},
	}, nil
}

// processFulfillmentTargetSearch shortlists the kiosks that could supply a basket.
//
// The response says advisory in as many words. It takes no lock, so a target reported able to supply
// may be emptied by another sale before the customer picks it; only reserving secures anything, and
// a client that treated this as a guarantee would promise goods it cannot deliver.
func (this *SalesOrderCrudApplicationServiceImpl) runFulfillmentTargetSearch(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	items, vErrs := readAvailabilityItems(params)
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	options, err := services.SearchFulfillmentTargets(
		ctx, readStringParam(params, paramOrgId), items, this.fulfillmentReservations)
	if err != nil {
		return nil, err
	}

	targets := make([]map[string]any, 0, len(options))
	for _, option := range options {
		shortages := make([]map[string]any, 0, len(option.Shortages))
		for _, shortage := range option.Shortages {
			shortages = append(shortages, map[string]any{
				"product_variant_id": shortage.ProductVariantId,
				"requested":          shortage.Requested,
				"available":          shortage.Available,
			})
		}
		targets = append(targets, map[string]any{
			"sales_point_id":        option.SalesPointId,
			"inventory_location_id": option.InventoryLocation,
			"can_fulfill_all":       option.CanFulfillAll,
			"shortages":             shortages,
		})
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"targets": targets,

			// Stated in the payload rather than only in documentation: a client reading this without
			// having read the CR must still learn that the answer secures nothing.
			"advisory": true,
		},
	}, nil
}

// readAvailabilityItems parses the basket. A malformed quantity is a violation rather than a silent
// zero: zero would report every kiosk able to supply nothing at all, which reads as success.
func readAvailabilityItems(
	params dmodel.DynamicFields,
) ([]itExt.AvailabilityItem, *ft.ClientErrors) {
	raw, present := params[paramItems]
	if !present || raw == nil {
		return nil, itemsViolation(reasonItemsMalformed,
			"name the products to check, as a list of {product_variant_id, quantity}")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, itemsViolation(reasonItemsMalformed,
			"items must be a list of {product_variant_id, quantity}")
	}

	items := make([]itExt.AvailabilityItem, 0, len(list))
	for _, entry := range list {
		fields, ok := entry.(map[string]any)
		if !ok {
			return nil, itemsViolation(reasonItemsMalformed,
				"each item must be an object with product_variant_id and quantity")
		}

		variantId, _ := fields[paramProductVariantId].(string)
		if variantId == "" {
			return nil, itemsViolation(reasonItemsMalformed,
				"each item must name a product_variant_id")
		}

		quantity, ok := readDecimalValue(fields[paramQuantity])
		if !ok || !quantity.IsPositive() {
			return nil, itemsViolation(reasonQuantityMalformed,
				"item '"+variantId+"' must carry a positive quantity")
		}

		items = append(items, itExt.AvailabilityItem{
			ProductVariantId: variantId,
			Quantity:         quantity,
		})
	}
	return items, nil
}

// The order's refund actions. Creating refunds reuses the update permission -- it changes what the
// sale owes, the same power as repricing it -- while the view is an ordinary read.

func (this *SalesOrderCrudApplicationServiceImpl) CreateRefunds(
	ctx corectx.Context, cmd itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionUpdate, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCreateOrderRefund(ctx, cmd)
}

func (this *SalesOrderCrudApplicationServiceImpl) ViewRefunds(
	ctx corectx.Context, query itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionRead, query); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runViewOrderRefunds(ctx, query)
}

// ListBills answers every bill of one order, superseded ones included.
//
// It exists for the client that lost the response to a confirm and knows only the order, and for
// reconciliation afterwards. Addressed by the ORDER rather than filtered on the bill collection, so
// the org check is the order's: a caller who may not read the sale may not enumerate what it owes.
func (this *SalesOrderCrudApplicationServiceImpl) ListBills(
	ctx corectx.Context, query itOrder.OrderActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, composable.PermissionRead, query); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runListOrderBills(ctx, query)
}

func (this *SalesOrderCrudApplicationServiceImpl) runListOrderBills(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	orderId := readStringParam(params, paramRecordId)

	// Cancelled bills included: this is the history, and a split that superseded a bill does not
	// unmake the payment somebody recorded against it.
	bills, err := services.BillsOfOrder(ctx, orderId, true)
	if err != nil {
		return nil, err
	}

	items := make([]any, 0, len(bills))
	for _, bill := range bills {
		items = append(items, bill)
	}
	return &dyn.OpResult[any]{HasData: true, Data: map[string]any{
		"sales_order_id": orderId,
		"items":          items,
	}}, nil
}

// Parameter names and the violation shape the refund bodies read.
const (
	paramRefundLines      = "lines"
	paramRefundReason     = "reason"
	reasonRefundMalformed = "sales_return.lines_malformed"
)

// processCreateOrderRefund raises a refund and reports it as PENDING.
//
// Creation is not success, and the response says so in as many words: the money has not moved yet,
// and a client that read a 200 as "refunded" would tell a customer they had been paid back when the
// legs may still be in flight.
func (this *SalesOrderCrudApplicationServiceImpl) runCreateOrderRefund(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	lines, vErrs := readRefundLines(params)
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	result, vErrs, err := services.CreateReturn(ctx, services.CreateReturnParams{
		SalesOrderId: readStringParam(params, paramRecordId),
		Reason:       readStringParam(params, paramRefundReason),

		// A customer asking is the only reason reachable from here. A fulfillment-failure refund is
		// raised by the failure policy itself and never by a request, which is what stops a caller
		// dressing an ordinary refund up as an automatic one to skip the goods-return step.
		RefundReason: models.SalesRefundReasonCustomerRequested,
		Lines:        lines,
	}, this.orderLock, services.ResolveSalesPolicy(ctx, this.effectiveSettings))
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"refund_id":     result.SalesReturnId,
			"refund_status": result.RefundStatus,
			"refund_total":  result.RefundTotal,
		},
	}, nil
}

// processViewOrderRefunds lists what has been asked for and what has actually been paid.
func (this *SalesOrderCrudApplicationServiceImpl) runViewOrderRefunds(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	views, err := services.ViewOrderRefunds(ctx, readStringParam(params, paramRecordId))
	if err != nil {
		return nil, err
	}

	refunds := make([]map[string]any, 0, len(views))
	for _, view := range views {
		items := make([]map[string]any, 0, len(view.Items))
		for _, item := range view.Items {
			items = append(items, map[string]any{
				"sales_order_line_id": item.SalesOrderLineId,
				"fulfillment_id":      item.FulfillmentId,
				"fulfillment_item_id": item.FulfillmentItemId,
				"requested_qty":       item.RequestedQty,

				// What actually went back, which is the number a customer service agent needs: the
				// requested figure says only what was asked for.
				"refunded_qty": item.RefundedQty,
			})
		}
		refunds = append(refunds, map[string]any{
			"refund_id":     view.RefundId,
			"refund_status": view.RefundStatus,
			"refund_reason": view.RefundReason,
			"return_type":   view.ReturnType,
			"refund_total":  view.RefundTotal,
			"items":         items,
		})
	}
	return &dyn.OpResult[any]{
		HasData: true,
		Data:    map[string]any{"refunds": refunds},
	}, nil
}

// readRefundLines parses what to refund. A line names an order line and how much of it; the
// fulfillment linkage is resolved server-side, so a caller cannot point a refund at somebody else's
// delivery.
func readRefundLines(
	params dmodel.DynamicFields,
) ([]services.CreateReturnLine, *ft.ClientErrors) {
	raw, present := params[paramRefundLines]
	if !present || raw == nil {
		return nil, refundViolation("name the lines to refund")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, refundViolation("lines must be a list of objects")
	}

	lines := make([]services.CreateReturnLine, 0, len(list))
	for _, entry := range list {
		fields, ok := entry.(map[string]any)
		if !ok {
			return nil, refundViolation("each line must be an object")
		}

		lineId, _ := fields["sales_order_line_id"].(string)
		if lineId == "" {
			return nil, refundViolation("each line must name a sales_order_line_id")
		}
		quantity, ok := readDecimalValue(fields[paramQuantity])
		if !ok || !quantity.IsPositive() {
			return nil, refundViolation("line " + lineId + " must carry a positive quantity")
		}

		lines = append(lines, services.CreateReturnLine{
			SalesOrderLineId: lineId,
			Quantity:         quantity,
			RequestedQty:     quantity,
		})
	}
	return lines, nil
}

func refundViolation(message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(
		models.SalesReturnSchemaName, reasonRefundMalformed, message))
	return vErrs
}

// readOptionalDateTimeParam reads an RFC 3339 UTC timestamp, or nil when absent.
func readOptionalDateTimeParam(params dmodel.DynamicFields, field string) (*time.Time, *ft.ClientErrors) {
	raw := readStringParam(params, field)
	if raw == "" {
		return nil, nil
	}
	parsed, err := model.ParseModelDateTime(raw)
	if err != nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, "sales_order."+field+"_malformed",
			"'"+field+"' must be an RFC 3339 UTC timestamp ending in Z"))
		return nil, vErrs
	}
	goTime := parsed.GoTime().UTC()
	return &goTime, nil
}
