package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
	itChannel "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/channel"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// apply_voucher hangs off sales_order, not the voucher code: it modifies the order while the code
// is untouched master data, so a till does not need write permission over campaign configuration to
// take a discount at the counter.

const (
	// Matches the action code seeded in 1007006_sales_voucher_iam.sql.
	PermissionApplyVoucher = "apply_voucher"

	ActionApplyVoucher = "apply_voucher"

	// Rides on `read`: the explanation is the same numbers with their provenance attached, and a
	// separate permission would let a role see a total it could not account for.
	PermissionExplainPrice = "read"

	ActionExplainPrice = "explain_price"

	// `update`: repricing changes the order's stored totals.
	PermissionReprice = "update"

	ActionReprice = "reprice"

	// The same power as the built-in POST route; create_order exists alongside it because it also
	// derives the channel, enforces idempotency and prices the result.
	PermissionCreateOrder = "create"

	ActionCreateOrder = "create_order"

	// Separate from `update`: confirming commits the business to a price and redeems vouchers,
	// cancelling unwinds a sale.
	PermissionConfirm = "confirm"
	PermissionCancel  = "cancel"

	ActionConfirmOrder = "confirm"
	ActionCancelOrder  = "cancel"

	// Split and merge hang off the bill, not the order: they restructure settlement units, and
	// dividing a bill is not the power to confirm or cancel the sale.
	PermissionSplitBill = "split"
	PermissionMergeBill = "merge"

	ActionSplitBill = "split"
	ActionMergeBill = "merge"

	// Its own power, not `update`: changing what a customer pays for reasons outside the price list
	// is a separate authority from correcting a quantity.
	PermissionManualDiscount = "manual_discount"

	ActionGrantManualDiscount  = "manual_discount"
	ActionRevokeManualDiscount = "revoke_manual_discount"

	PermissionPayBill    = "pay"
	PermissionSettleBill = "settle"

	ActionPayBill    = "pay"
	ActionSettleBill = "settle"
)

func defineSalesBillActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSplitBill,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/split",
			Permission:  PermissionSplitBill,
			MainProcess: processSplitBill,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionPayBill,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/pay",
			Permission:  PermissionPayBill,
			MainProcess: processPayBill,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionSettleBill,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/settle",
			Permission:  PermissionSettleBill,
			MainProcess: processSettleBill,
		}),
		defineGatewayPaymentActions(engine),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionMergeBill,
			ActionType: drif.ActionTypeGeneric,
			// Collection-level: a merge names several bills and produces a new one, so the bills
			// travel in the body.
			RestPath:    "merge",
			Permission:  PermissionMergeBill,
			MainProcess: processMergeBills,
		}),
	)
}

func processSplitBill(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.SplitBill(ctx, services.SplitBillParams{
		SourceBillId: readStringParam(input.Params, paramId),
		Parts:        readSplitParts(input.Params),
	}, orderLock, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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

func processMergeBills(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	result, vErrs, err := services.MergeBills(ctx, services.MergeBillParams{
		SourceBillIds: readStringsParam(input.Params, "source_bill_ids"),
	}, orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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

func readStringsParam(params map[string]any, field string) []string {
	raw, ok := params[field].([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(raw))
	for _, item := range raw {
		if typed, ok := item.(string); ok {
			values = append(values, typed)
		}
	}
	return values
}

const (
	paramVoucherCode = "code"
	paramNowUnix     = "now_unix"
)

func defineSalesOrderVoucherActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionApplyVoucher,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/apply_voucher",
			Permission:  PermissionApplyVoucher,
			MainProcess: processApplyVoucher,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionCreateOrder,
			ActionType: drif.ActionTypeGeneric,
			// Collection-level: there is no record yet to name.
			RestPath:    "create_order",
			Permission:  PermissionCreateOrder,
			MainProcess: processCreateOrder,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReprice,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/reprice",
			Permission:  PermissionReprice,
			MainProcess: processReprice,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfirmOrder,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/confirm",
			Permission:  PermissionConfirm,
			MainProcess: processConfirmOrder,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancelOrder,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancel,
			MainProcess: processCancelOrder,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionExplainPrice,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/explain_price",
			Permission:  PermissionExplainPrice,
			MainProcess: processExplainPrice,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionGrantManualDiscount,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/manual_discount",
			Permission:  PermissionManualDiscount,
			MainProcess: processGrantManualDiscount,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionRevokeManualDiscount,
			ActionType: drif.ActionTypeGeneric,
			RestPath:   ":id/revoke_manual_discount",

			// Revoking takes the same permission as granting, not `update`: withdrawing a discount
			// raises what the customer pays.
			Permission:  PermissionManualDiscount,
			MainProcess: processRevokeManualDiscount,
		}),
	)
}

// processExplainPrice answers why an order costs what it costs. A POST only because that is what
// ActionTypeGeneric routes are here; it writes nothing and is safely repeatable.
func processExplainPrice(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	orderId := readStringParam(input.Params, paramId)

	explanation, err := services.ExplainOrderPrice(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if explanation == nil {
		return &drif.ActionResult{
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

	return &drif.ActionResult{
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

// processApplyVoucher is transport only; every rule lives in services.ApplyVoucher so the operation
// stays reachable from CQRS and from another module's port without HTTP.
func processApplyVoucher(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	orderId := readStringParam(input.Params, paramId)

	order, err := services.LoadSalesOrderForVoucher(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		// A client error rather than a fault: the caller named a record that does not exist.
		return &drif.ActionResult{
			ClientErrors: *services.OrderNotFoundErrors(orderId),
		}, nil
	}

	// The clock is read once at the edge and passed inward, so a test can reproduce it and two gates
	// in the same request cannot disagree about the time.
	nowUnix := readInt64Param(input.Params, paramNowUnix)
	if nowUnix == 0 {
		nowUnix = services.NowUnix()
	}

	result, vErrs, err := services.ApplyVoucher(ctx, services.ApplyVoucherParams{
		Code:              readStringParam(input.Params, paramVoucherCode),
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
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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

// Ports set once at Init. Package vars because an action callback is handed only its own engine.
var (
	taxCalculation    itExt.TaxCalculationExtService
	productVariants   itExt.ProductVariantExtService
	pricingBasis      itExt.ProductPricingBasisExtService
	effectiveSettings itExt.EffectiveSettingsExtService
	orderFulfillment  itExt.FulfillmentExtService

	// orderLock guards the operations that are not single-row updates. Confirm and cancel refuse to
	// run without it rather than proceeding unguarded.
	orderLock lock.DistributedLock

	// channelPayments answers "does this channel accept this method"; the second gate, "is the method
	// usable at all", is paymentMethods.
	channelPayments itChannel.ChannelPaymentAppService
)

// SetPricingPorts must be called by Init before any request.
func SetPricingPorts(
	tax itExt.TaxCalculationExtService,
	settings itExt.EffectiveSettingsExtService,
	dLock lock.DistributedLock,
	products itExt.ProductVariantExtService,
	fulfillment itExt.FulfillmentExtService,
	basis itExt.ProductPricingBasisExtService,
) {
	taxCalculation = tax
	effectiveSettings = settings
	orderLock = dLock

	// The product port gates "nothing withdrawn from sale may be ordered". Nil is supported: a
	// deployment without inventory has no such master, so the gate permits rather than refusing.
	productVariants = products

	// The fulfilment port asks Inventory to hold the goods a confirm has sold. Nil is supported: the
	// request is written and left pending rather than refused, because the sale is real.
	orderFulfillment = fulfillment

	// The pricing-basis port re-reads a product's base price and cost on every reprice. Nil is
	// supported: pricing then falls back to the price already stored on the line.
	pricingBasis = basis
}

// SetChannelPaymentService must be called AFTER the application services are built, not with the
// other ports: ChannelPaymentAppService is one of Sales' own services, registered several steps
// after external ports, and resolving it eagerly in InitExternal is a same-module cycle that fails
// the boot with "missing type: channel.ChannelPaymentAppService".
func SetChannelPaymentService(service itChannel.ChannelPaymentAppService) {
	channelPayments = service
}

// processReprice is a separate action rather than a hook in the line CRUD routes, so a caller adding
// three lines gets one reprice rather than three adjustment chains nobody saw. A caller can forget
// it, leaving stale draft totals, but confirm reprices unconditionally before freezing anything.
func processReprice(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	orderId := readStringParam(input.Params, paramId)

	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.RepriceOrder(ctx, orderId, taxCalculation, policy, pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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

func processCreateOrder(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	params := services.CreateOrderParams{
		SalesChannelCode:  readStringParam(input.Params, "sales_channel_code"),
		SalesPointId:      readStringParam(input.Params, "sales_point_id"),
		CustomerReference: readStringParam(input.Params, "customer_reference"),
		CurrencyCode:      readStringParam(input.Params, "currency_code"),
		ExternalReference: readStringParam(input.Params, "external_reference"),
		IdempotencyKey:    readStringParam(input.Params, "idempotency_key"),
		Lines:             readOrderLines(input.Params),
	}

	result, vErrs, err := services.CreateOrder(ctx, params, taxCalculation, productVariants, pricingBasis, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
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

	return &drif.ActionResult{HasData: true, Data: data}, nil
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
		})
	}
	return lines
}

// readDecimalParam accepts whatever shape JSON delivered. A decimal crosses as a string so it does
// not lose precision; a float64 is accepted so a bare number does not silently become zero.
func readDecimalParam(params map[string]any, field string) decimal.Decimal {
	value, ok := params[field]
	if !ok || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case string:
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	case int:
		return decimal.NewFromInt(int64(typed))
	case int64:
		return decimal.NewFromInt(typed)
	}
	return decimal.Zero
}

func processConfirmOrder(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.ConfirmOrder(ctx,
		readStringParam(input.Params, paramId), orderLock, taxCalculation, orderFulfillment, pricingBasis, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	data := map[string]any{
		"sales_order_id":       result.SalesOrderId,
		"status":               result.Status,
		"confirmed_at":         result.ConfirmedAt,
		"redeemed_voucher_ids": result.RedeemedVoucherIds,

		// In the response, not just the log: a kiosk that believed a confirm was complete would
		// dispense goods against an order with no bill and no fulfilment request.
		"pending": result.Pending,
	}
	if result.Pricing != nil {
		data["grand_total"] = result.Pricing.GrandTotal
	}

	return &drif.ActionResult{HasData: true, Data: data}, nil
}

func processCancelOrder(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	result, vErrs, err := services.CancelOrder(ctx,
		readStringParam(input.Params, paramId),
		readStringParam(input.Params, "reason"),
		orderLock)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
		HasData: true,
		Data: map[string]any{
			"sales_order_id":       result.SalesOrderId,
			"status":               result.Status,
			"cancelled_at":         result.CancelledAt,
			"released_voucher_ids": result.ReleasedVoucherIds,
			"pending":              result.Pending,
		},
	}, nil
}

// processPayBill records a payment and settles immediately rather than in a second request: a bill
// whose last payment just landed is settled by that fact.
func processPayBill(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)
	billId := readStringParam(input.Params, paramId)

	result, vErrs, err := services.RecordPayment(ctx, services.RecordPaymentParams{
		SalesBillId:           billId,
		PaymentMethodId:       readStringParam(input.Params, "payment_method_id"),
		Amount:                readDecimalParam(input.Params, "amount"),
		CurrencyCode:          readStringParam(input.Params, "currency_code"),
		ExternalTransactionId: readStringParam(input.Params, "external_transaction_id"),
		ProviderReference:     readStringParam(input.Params, "provider_reference"),
		Status:                readStringParam(input.Params, "status"),
	}, paymentMethods, channelPayments, policy)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
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
	return &drif.ActionResult{HasData: true, Data: data}, nil
}

// processSettleBill is exposed separately from pay so an operator can reconcile a bill whose
// payments arrived by a path Sales did not record.
func processSettleBill(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	billId := readStringParam(input.Params, paramId)

	result, vErrs, err := services.SettleBillIfPaid(ctx, billId)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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

// processGrantManualDiscount records an operator override and reprices.
func processGrantManualDiscount(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	result, vErrs, err := services.GrantManualDiscount(ctx, services.GrantManualDiscountParams{
		SalesOrderId:     readStringParam(input.Params, paramId),
		SalesOrderLineId: readStringParam(input.Params, "sales_order_line_id"),
		Amount:           readDecimalParam(input.Params, "discount_amount"),
		Reason:           readStringParam(input.Params, "reason"),
	}, taxCalculation, policy, pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	return &drif.ActionResult{
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
func processRevokeManualDiscount(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	policy := services.ResolveSalesPolicy(ctx, effectiveSettings)

	vErrs, err := services.RevokeManualDiscount(ctx,
		readStringParam(input.Params, paramId),
		readStringParam(input.Params, "sales_manual_discount_id"),
		taxCalculation, policy, pricingBasis)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}
	return &drif.ActionResult{
		HasData: true,
		Data:    map[string]any{"sales_order_id": readStringParam(input.Params, paramId)},
	}, nil
}
