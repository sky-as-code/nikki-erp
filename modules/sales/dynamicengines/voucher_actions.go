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

// The apply-voucher action (BR 71, SALES-023).
//
// It hangs off SALES_ORDER rather than off the voucher code, and that placement is the decision
// worth explaining. Applying a voucher modifies the ORDER — it reserves a use and changes what the
// basket costs — while the code itself is untouched master data. Putting it on the code would mean a
// till needed write permission over campaign configuration in order to take a discount at the
// counter, which is exactly backwards.
//
// The permission is therefore `apply_voucher` on `sales_order`: a power over this sale, held by
// whoever may edit this sale.

const (
	// PermissionApplyVoucher matches the action code seeded in 1007006_sales_voucher_iam.sql.
	PermissionApplyVoucher = "apply_voucher"

	ActionApplyVoucher = "apply_voucher"

	// PermissionExplainPrice rides on `read`, not a permission of its own.
	//
	// The explanation contains nothing the order itself does not already show - it is the same
	// numbers with their provenance attached. A separate permission would let a role see a total it
	// could not account for, which is the opposite of what BR 87.9 is for.
	PermissionExplainPrice = "read"

	ActionExplainPrice = "explain_price"

	// PermissionReprice is `update`: repricing changes the order's stored totals, which is exactly
	// the power an update represents. A separate permission would let a role change what a customer
	// owes without holding the permission to edit the order.
	PermissionReprice = "update"

	ActionReprice = "reprice"

	// PermissionCreateOrder is `create` on sales_order - the same power the built-in POST route
	// represents. The operation exists alongside that route rather than replacing it because it
	// does considerably more: it derives the channel, enforces idempotency and prices the result.
	PermissionCreateOrder = "create"

	ActionCreateOrder = "create_order"

	// Confirm and cancel are separate powers from `update`, and deliberately so: confirming commits
	// the business to a price and redeems vouchers, cancelling unwinds a sale. A role that may
	// correct a line should not thereby be able to do either.
	PermissionConfirm = "confirm"
	PermissionCancel  = "cancel"

	ActionConfirmOrder = "confirm"
	ActionCancelOrder  = "cancel"

	// Split and merge hang off the BILL, not the order: they restructure settlement units, and a
	// role that may divide a bill need not be able to confirm or cancel the sale itself.
	PermissionSplitBill = "split"
	PermissionMergeBill = "merge"

	ActionSplitBill = "split"
	ActionMergeBill = "merge"

	// PermissionManualDiscount is its own power, NOT `update` (BR 87.4). Changing what a customer
	// pays for reasons outside the price list is exactly the authority a requirement gates
	// separately: a role that may correct a quantity should not thereby be able to discount a sale.
	PermissionManualDiscount = "manual_discount"

	ActionGrantManualDiscount  = "manual_discount"
	ActionRevokeManualDiscount = "revoke_manual_discount"

	PermissionPayBill    = "pay"
	PermissionSettleBill = "settle"

	ActionPayBill    = "pay"
	ActionSettleBill = "settle"
)

// defineSalesBillActions adds split and merge to the bill engine.
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
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName: ActionMergeBill,
			ActionType: drif.ActionTypeGeneric,
			// Collection-level: a merge names several bills and produces a new one, so there is no
			// single record to hang it off. The bills travel in the body.
			RestPath:    "merge",
			Permission:  PermissionMergeBill,
			MainProcess: processMergeBills,
		}),
	)
}

// processSplitBill divides one bill into several (BR 37, SALES-025).
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

			// Both totals are returned so a caller can see BR 37's invariant held rather than
			// taking it on trust.
			"total_before": result.TotalBefore,
			"total_after":  result.TotalAfter,
		},
	}, nil
}

// processMergeBills combines several bills into one (BR 38, SALES-026).
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

// readSplitParts pulls the requested parts out of the request body.
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

// readStringsParam reads a list of ids from the request body.
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

// Params the action reads.
const (
	paramVoucherCode = "code"
	paramNowUnix     = "now_unix"
)

// defineSalesOrderVoucherActions adds the sales order's custom actions.
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
			// Collection-level: there is no record yet to name. The only other action in this module
			// without an :id is resolve, for the same reason.
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

			// Revoking takes the SAME permission as granting, not `update`. Withdrawing a discount
			// changes what the customer pays just as granting one does, and a role that could undo
			// an approved discount without holding the discount power could quietly raise a price
			// somebody had authorised lowering.
			Permission:  PermissionManualDiscount,
			MainProcess: processRevokeManualDiscount,
		}),
	)
}

// processExplainPrice answers why an order costs what it costs (BR 87.9, SALES-021).
//
// A POST despite reading nothing, because that is what ActionTypeGeneric routes are in this engine;
// it writes nothing and is safely repeatable.
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

			// Reported rather than hidden. A false value means the stored adjustments do not account
			// for the stored net, which is a bug worth surfacing on the screen that would otherwise
			// display an explanation that quietly does not add up.
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

// processApplyVoucher reads the request and hands it to the domain service.
//
// Transport only: it reads params, resolves the order's channel and point, and converts the answer.
// Every rule lives in services.ApplyVoucher, so the same operation is reachable from CQRS and from
// another module's port without going through HTTP (docs/wiki/07 §6.7).
func processApplyVoucher(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	orderId := readStringParam(input.Params, paramId)

	order, err := services.LoadSalesOrderForVoucher(ctx, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		// No such order. Reported as a client error rather than a fault: the caller named a record
		// that does not exist, which a 404 or a 400 describes and a 500 does not.
		return &drif.ActionResult{
			ClientErrors: *services.OrderNotFoundErrors(orderId),
		}, nil
	}

	// The clock is read HERE, at the edge, and passed inward. Every rule beneath this point takes
	// the instant as data, which is what makes applying a voucher reproducible in a test and what
	// stops two gates in the same request disagreeing about what time it is.
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

		// The basket facts come from the order's own lines. An empty basket still evaluates: a
		// program with no conditions applies to it, and one with a minimum spend does not.
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

// readInt64Param reads an optional integer param, or zero when it is absent.
//
// Every numeric shape is accepted for the same reason the settings readers accept them: a value that
// arrived as JSON is a float64, and a reader taking only int64 would silently ignore it.
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

// The tax port and the settings port the reprice action reads.
//
// Package vars set once at Init, for the same reason the payment method port is: an action callback
// is handed only its own engine, so a port has to reach it some other way.
var (
	taxCalculation    itExt.TaxCalculationExtService
	productVariants   itExt.ProductVariantExtService
	pricingBasis      itExt.ProductPricingBasisExtService
	effectiveSettings itExt.EffectiveSettingsExtService
	orderFulfillment  itExt.FulfillmentExtService

	// orderLock guards the operations that are not single-row updates (D-30). Confirm and cancel
	// both refuse to run without it rather than proceeding unguarded.
	orderLock lock.DistributedLock

	// channelPayments answers CR 33's first gate: does this channel accept this method. The second
	// gate - is the method usable at all - is paymentMethods, which the payment actions already hold.
	channelPayments itChannel.ChannelPaymentAppService
)

// SetPricingPorts installs the ports the pricing actions read. Init calls it before any request.
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

	// The product port gates BR 69: nothing withdrawn from sale may be ordered. Nil is a supported
	// value - a deployment without inventory has no master to be withdrawn from - and the gate
	// permits in that case rather than refusing, which would make Sales unusable there.
	productVariants = products

	// The fulfilment port asks Inventory to hold the goods a confirm has sold (SALES-049). Nil is
	// supported for the same reason as the product port: a deployment without inventory has no
	// stock to reserve. The request is then written and left pending rather than refused, because
	// the sale is real and the goods are genuinely owed.
	orderFulfillment = fulfillment

	// The pricing-basis port re-reads a product's base price and cost on every reprice. Nil is
	// supported, like the two above: without inventory there is no product to re-read, and pricing
	// falls back to the price already stored on the line rather than refusing to reprice at all.
	pricingBasis = basis
}

// SetChannelPaymentService installs the mapping gate, and is called AFTER the application services
// are built rather than with the other ports.
//
// The ordering is forced and worth stating. Init binds external ports FIRST, because a derived
// service resolves its ports when it is constructed - but ChannelPaymentAppService is one of Sales'
// OWN application services, registered several steps later. Resolving it eagerly in InitExternal is
// a same-module cycle: the boot fails with "missing type: channel.ChannelPaymentAppService", which
// is exactly what happened when this was first written that way.
func SetChannelPaymentService(service itChannel.ChannelPaymentAppService) {
	channelPayments = service
}

// processReprice recomputes a draft order after its lines changed (BR 70, SALES-011).
//
// A separate action rather than a hook inside the line CRUD routes, and that is a deliberate choice
// worth stating: a caller adding three lines wants ONE reprice, not three. Repricing on every line
// write would run the engine twice for nothing and produce two adjustment chains that were never
// shown to anybody.
//
// The cost is that a caller can forget to call it, leaving a draft whose totals are stale. Confirm
// (SALES-013) reprices unconditionally before it freezes anything, so a forgotten reprice cannot
// reach a customer — it can only make a draft screen briefly wrong.
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

// processCreateOrder creates a draft order and prices it (BR 69, SALES-012).
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

		// True on the idempotent replay path. The caller gets a success either way (D-29); this
		// says whether anything was actually written, which is what a log or a metric wants.
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

// readOrderLines pulls the requested lines out of the request body.
//
// A line whose shape is not a map is skipped rather than erroring. The validation that follows
// refuses an empty basket that should not have been empty, and reporting "lines[2] is not an object"
// is a parser complaint rather than a business one.
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

// readDecimalParam reads a money or quantity value, in whatever shape JSON delivered it.
//
// A decimal crosses JSON as a string so it does not lose precision; a float64 is accepted because a
// caller that sent a bare number should not silently order zero.
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

// processConfirmOrder commits a draft order (BR 72, SALES-013).
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

		// Named in the response rather than logged. A kiosk that believed a confirm was complete
		// will dispense goods against an order with no bill and no fulfilment request.
		"pending": result.Pending,
	}
	if result.Pricing != nil {
		data["grand_total"] = result.Pricing.GrandTotal
	}

	return &drif.ActionResult{HasData: true, Data: data}, nil
}

// processCancelOrder cancels an order if its state allows it (BR 43, SALES-014).
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

// processPayBill records a payment against a bill (BR 75, SALES-027).
//
// Settlement follows immediately rather than being a second request: a bill whose last payment just
// landed is settled by that fact, and leaving it open would need somebody to press a button on a
// bill with nothing owed.
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

		// Reported separately, never folded into the captured total: change is handed back, so
		// counting it would overstate what the sale was worth (BR 42).
		"change_due":      result.ChangeDue,
		"already_existed": result.AlreadyExisted,
	}
	if settled != nil {
		data["bill_status"] = settled.Status
		data["payment_status"] = settled.PaymentStatus
	}
	return &drif.ActionResult{HasData: true, Data: data}, nil
}

// processSettleBill closes a bill whose money is fully in (BR 76, SALES-028).
//
// Exposed separately from pay so an operator can reconcile a bill whose payments arrived by some
// path Sales did not record - a bank transfer matched by hand, say.
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

// processGrantManualDiscount records an operator override and reprices (BR 87.4, SALES-039).
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

			// BOTH totals, because the engine caps a discount at what is owed: the difference is
			// not always the amount asked for, and an operator told only "granted" would not know.
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
