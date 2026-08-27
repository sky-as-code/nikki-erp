package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Creating a sales order (BR 69, CR 21-22 and 56-57, SALES-012).
//
// Two rules here are load-bearing and neither is obvious from the field list.
//
// # The channel is DERIVED, never accepted (D-20, CR 57)
//
// A request may name a channel code, and the sales point also names a channel. The stored channel is
// always the SALES POINT's, and a supplied code that disagrees is rejected outright. This is the
// anti-spoofing rule: a kiosk claiming the `vdmc` channel while passing an eCommerce sales point id
// must fail, because otherwise it could book itself into whichever channel had the better prices or
// the laxer payment rules. Deriving it makes CR 20's invariant true by construction rather than by a
// check somebody could forget to run.
//
// # A duplicate idempotency key returns the ORIGINAL order, successfully (D-29)
//
// Not a conflict, not an error. BR 7.2's "a machine retry must not create two orders" is only true if
// the retry gets a success back - a gateway or kiosk told "409" retries forever, and each retry is
// another chance to create the duplicate the rule exists to prevent. The unique index is the
// mechanism; returning the existing row is the contract.

// CreateOrderParams is what creating an order needs.
type CreateOrderParams struct {
	// SalesChannelCode is optional and is only ever CHECKED, never used to decide the channel.
	// See the anti-spoofing note above.
	SalesChannelCode string

	SalesPointId string

	CustomerReference string
	CurrencyCode      string

	Lines []CreateOrderLine

	// ExternalReference is the caller's own identifier for this sale, unique per channel.
	ExternalReference string

	// IdempotencyKey makes a retry safe. Unique per channel when present; absent means the caller
	// accepts that a retry creates a second order.
	IdempotencyKey string

	OrgId string
}

// CreateOrderLine is one requested line.
type CreateOrderLine struct {
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal

	// UnitPrice is the catalogue price the caller resolved. The engine uses it as the fallback when
	// no pricelist item matches.
	UnitPrice decimal.Decimal

	ProductCode string
	ProductName string
}

// CreateOrderResult is the created order plus what it priced to.
type CreateOrderResult struct {
	SalesOrderId   string
	OrderNumber    string
	SalesChannelId string

	Pricing *RepriceResult

	// AlreadyExisted marks the idempotent replay path. The caller returns success either way; this
	// says whether anything was actually written, which is what a log or a metric wants to know.
	AlreadyExisted bool
}

// The refusal reasons create can produce.
const (
	ReasonChannelMismatch     = "sales_order.sales_channel_mismatch"
	ReasonPointNotFound       = "sales_order.sales_point_not_found"
	ReasonPointNotSellable    = "sales_order.sales_point_not_sellable"
	ReasonChannelNotSellable  = "sales_order.sales_channel_not_sellable"
	ReasonQuantityNotPositive = "sales_order.quantity_not_positive"
	ReasonVariantMissing      = "sales_order.product_variant_missing"
)

// CreateOrder validates and writes a new draft order, then prices it.
func CreateOrder(
	ctx corectx.Context,
	params CreateOrderParams,
	taxSvc itExt.TaxCalculationExtService,
	products itExt.ProductVariantExtService,
	policy SalesPolicy,
) (*CreateOrderResult, *ft.ClientErrors, error) {
	point, channel, vErrs, err := resolveSellingPlace(ctx, params)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	channelId := stringOf(channel, models.SalesChannelFieldId)

	if vErrs := assertLinesRequestable(params.Lines); vErrs != nil {
		return nil, vErrs, nil
	}

	// BR 69: nothing withdrawn from sale may be ordered. Checked BEFORE the idempotency replay and
	// before any write, because an order naming an unsellable variant must not exist at all — one
	// written and then rejected would still be a row a till could read and a customer could be
	// charged against.
	sellableErrs, err := assertVariantsSellable(ctx, params.Lines, products)
	if err != nil {
		return nil, nil, err
	}
	if sellableErrs != nil {
		return nil, sellableErrs, nil
	}

	// The idempotent replay, checked BEFORE writing anything. The unique index is what makes this
	// safe under a race - two simultaneous retries cannot both pass this check, and the loser's
	// insert fails on the index rather than creating a second order.
	if params.IdempotencyKey != "" {
		existing, err := findOrderByIdempotencyKey(ctx, channelId, params.IdempotencyKey)
		if err != nil {
			return nil, nil, err
		}
		if existing != nil {
			return &CreateOrderResult{
				SalesOrderId:   stringOf(existing, models.SalesOrderFieldId),
				OrderNumber:    stringOf(existing, models.SalesOrderFieldOrderNumber),
				SalesChannelId: channelId,
				AlreadyExisted: true,
			}, nil, nil
		}
	}

	orderId, orderNumber, err := writeDraftOrder(ctx, params, point, channelId)
	if err != nil {
		// A collision on the idempotency index means a concurrent retry won the race. That is the
		// success path, not a failure: re-read and return what the winner created (D-29).
		if isUniqueViolation(err) && params.IdempotencyKey != "" {
			existing, lookupErr := findOrderByIdempotencyKey(ctx, channelId, params.IdempotencyKey)
			if lookupErr != nil {
				return nil, nil, lookupErr
			}
			if existing != nil {
				return &CreateOrderResult{
					SalesOrderId:   stringOf(existing, models.SalesOrderFieldId),
					OrderNumber:    stringOf(existing, models.SalesOrderFieldOrderNumber),
					SalesChannelId: channelId,
					AlreadyExisted: true,
				}, nil, nil
			}
		}
		return nil, nil, err
	}

	// Priced after the lines are stored, because the engine reads them back. BR 69 requires a total
	// to have been computed before one is returned, so a create that could not price is a create
	// that failed.
	priced, vErrs, err := RepriceOrder(ctx, orderId, taxSvc, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	return &CreateOrderResult{
		SalesOrderId:   orderId,
		OrderNumber:    orderNumber,
		SalesChannelId: channelId,
		Pricing:        priced,
	}, nil, nil
}

// resolveSellingPlace loads the sales point, derives its channel, and checks both may sell.
//
// The order matters: the point is loaded first because it is what decides the channel. Loading the
// channel from the request and then checking the point against it would be the same code with the
// trust running the wrong way.
func resolveSellingPlace(
	ctx corectx.Context, params CreateOrderParams,
) (point dmodel.DynamicFields, channel dmodel.DynamicFields, vErrs *ft.ClientErrors, err error) {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		errs := ft.NewClientErrors()
		errs.Append(*ft.NewBusinessViolation(field, reason, message))
		return errs
	}

	point, err = loadRecord(ctx,
		models.SalesPointSchemaName, models.SalesPointFieldId, params.SalesPointId)
	if err != nil {
		return nil, nil, nil, err
	}
	if point == nil {
		return nil, nil, refuse("sales_point_id", ReasonPointNotFound,
			"no sales point exists with id '"+params.SalesPointId+"'"), nil
	}

	// The channel comes from the POINT. This single line is the anti-spoofing rule.
	channelId := stringOf(point, models.SalesPointFieldSalesChannelId)
	channel, err = loadRecord(ctx,
		models.SalesChannelSchemaName, models.SalesChannelFieldId, channelId)
	if err != nil {
		return nil, nil, nil, err
	}
	if channel == nil {
		// A point whose channel is gone cannot sell. Reported against the point, because that is
		// the record the caller named and the one an administrator would go and fix.
		return nil, nil, refuse("sales_point_id", ReasonChannelNotSellable,
			"this sales point references a sales channel that no longer exists"), nil
	}

	// A supplied code is CHECKED against the derived channel, never used instead of it (CR 57).
	if params.SalesChannelCode != "" &&
		params.SalesChannelCode != stringOf(channel, models.SalesChannelFieldCode) {
		return nil, nil, refuse("sales_channel_code", ReasonChannelMismatch,
			"the sales channel code does not match the channel this sales point belongs to"), nil
	}

	if !canSell(channel, models.SalesChannelFieldStatus,
		string(models.SalesChannelStatusActive)) {
		return nil, nil, refuse("sales_point_id", ReasonChannelNotSellable,
			"the sales channel is not active"), nil
	}
	if !canSell(point, models.SalesPointFieldStatus,
		string(models.SalesPointStatusActive)) {
		return nil, nil, refuse("sales_point_id", ReasonPointNotSellable,
			"the sales point is not active"), nil
	}
	return point, channel, nil, nil
}

// canSell reports whether a record is both active and unarchived.
//
// Both gates, because they mean different things: archived is retired for good, suspended is stopped
// for now, and either one must prevent a new sale.
func canSell(record dmodel.DynamicFields, statusField, activeStatus string) bool {
	if boolOf(record, basemodel.FieldIsArchived) {
		return false
	}
	return stringOf(record, statusField) == activeStatus
}

// assertLinesRequestable checks what can be checked without reading another module.
//
// Variant existence and sellability are NOT checked here. They need inventory's product port, which
// Sales does not yet bind - see the SALES-012 note in 02-progress.md. A line naming a variant that
// does not exist is therefore stored and will fail later, which is worse than refusing it now and is
// recorded rather than hidden.
func assertLinesRequestable(lines []CreateOrderLine) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	for index, line := range lines {
		field := "lines[" + decimal.NewFromInt(int64(index)).String() + "]"

		if !line.Quantity.IsPositive() {
			vErrs.Append(*ft.NewBusinessViolation(field, ReasonQuantityNotPositive,
				"a line must order more than zero; a line ordering nothing is not a line"))
		}
		if line.ProductVariantId == "" {
			vErrs.Append(*ft.NewBusinessViolation(field, ReasonVariantMissing,
				"a line must name a product variant"))
		}
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return vErrs
}

// assertVariantsSellable refuses an order naming a variant Inventory has withdrawn (BR 69).
//
// # Why this is a separate step from assertLinesRequestable
//
// That one is pure — it asks whether the request is well-formed, and needs nothing but the request.
// This one asks another module a question, so folding them together would make the pure check
// impossible to test without a container, and every caller pay for a round trip to learn that a
// quantity was negative.
//
// # A nil port PERMITS rather than refuses
//
// The port is bound in every build that ships inventory, so nil means a deployment without it —
// and in such a deployment there is no master to be withdrawn from. Refusing would make Sales
// unusable there rather than safe. Note this is the opposite reading from the tax port, which fails
// CLOSED: an unresolved tax silently undercharges the business, while an unchecked variant is a
// question nobody in that deployment can answer.
func assertVariantsSellable(
	ctx corectx.Context, lines []CreateOrderLine, products itExt.ProductVariantExtService,
) (*ft.ClientErrors, error) {
	if products == nil {
		return nil, nil
	}

	variantIds := make([]string, 0, len(lines))
	for _, line := range lines {
		if line.ProductVariantId != "" {
			variantIds = append(variantIds, line.ProductVariantId)
		}
	}
	if len(variantIds) == 0 {
		return nil, nil
	}

	result, err := products.AssertSellable(ctx, itExt.AssertSellableQuery{
		ProductVariantIds: variantIds,
	})
	if err != nil {
		return nil, err
	}
	if result == nil || len(result.NotSellable) == 0 {
		return nil, nil
	}

	// Reported per LINE rather than per variant, because the caller sent lines and fixing the order
	// means editing one. A refusal naming only the variant id would leave an operator scanning a
	// twenty-line basket for it.
	vErrs := ft.NewClientErrors()
	for index, line := range lines {
		reason, refused := result.NotSellable[line.ProductVariantId]
		if !refused {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(
			"lines["+decimal.NewFromInt(int64(index)).String()+"]",
			reason,
			"product variant '"+line.ProductVariantId+"' cannot be sold"))
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return vErrs, nil
}

// findOrderByIdempotencyKey looks for an order already created under this key.
func findOrderByIdempotencyKey(
	ctx corectx.Context, channelId, key string,
) (dmodel.DynamicFields, error) {
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.SalesOrderFieldSalesChannelId, dmodel.Equals, channelId))
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.SalesOrderFieldIdempotencyKey, dmodel.Equals, key))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return found.Data.Items[0], nil
}

// writeDraftOrder inserts the order and its lines, in one transaction.
//
// Together, because an order whose lines failed to write is an empty order that looks deliberate -
// and an empty draft is a legitimate state (BR 69), so nothing downstream would flag it.
func writeDraftOrder(
	ctx corectx.Context, params CreateOrderParams, point dmodel.DynamicFields, channelId string,
) (orderId string, orderNumber string, err error) {
	id, err := model.NewId()
	if err != nil {
		return "", "", err
	}
	orderId = string(*id)
	orderNumber = "SO-" + orderId

	orgId := params.OrgId
	if orgId == "" {
		orgId = stringOf(point, basemodel.FieldOrgId)
	}

	err = withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		orderEngine, err := engineFor(models.SalesOrderSchemaName)
		if err != nil {
			return err
		}

		fields := dmodel.DynamicFields{
			models.SalesOrderFieldId:             orderId,
			models.SalesOrderFieldOrderNumber:    orderNumber,
			models.SalesOrderFieldSalesChannelId: channelId,
			models.SalesOrderFieldSalesPointId:   params.SalesPointId,
			models.SalesOrderFieldCurrencyCode:   params.CurrencyCode,

			models.SalesOrderFieldStatus:            string(models.SalesOrderStatusDraft),
			models.SalesOrderFieldPaymentStatus:     string(models.SalesOrderPaymentStatusUnpaid),
			models.SalesOrderFieldFulfillmentStatus: string(models.SalesOrderFulfillmentStatusPending),
			models.SalesOrderFieldInvoiceStatus:     string(models.SalesOrderInvoiceStatusNotRequested),

			// Zeroed rather than omitted. The reprice that follows overwrites them, and a NOT NULL
			// column with no value would fail the insert before it got there.
			models.SalesOrderFieldSubtotal:      decimal.Zero,
			models.SalesOrderFieldDiscountTotal: decimal.Zero,
			models.SalesOrderFieldTaxTotal:      decimal.Zero,
			models.SalesOrderFieldGrandTotal:    decimal.Zero,

			basemodel.FieldOrgId: orgId,
		}
		if params.CustomerReference != "" {
			fields[models.SalesOrderFieldCustomerReference] = params.CustomerReference
		}
		if params.ExternalReference != "" {
			fields[models.SalesOrderFieldExternalReference] = params.ExternalReference
		}
		if params.IdempotencyKey != "" {
			fields[models.SalesOrderFieldIdempotencyKey] = params.IdempotencyKey
		}

		if _, err := orderEngine.ResourceRepository().Insert(tranxCtx, fields); err != nil {
			return err
		}
		return writeOrderLines(tranxCtx, orderId, orgId, params.Lines)
	})
	if err != nil {
		return "", "", err
	}
	return orderId, orderNumber, nil
}

// writeOrderLines inserts the requested lines, numbered from one.
func writeOrderLines(
	ctx corectx.Context, orderId, orgId string, lines []CreateOrderLine,
) error {
	if len(lines) == 0 {
		// An order with zero lines is a valid draft (BR 69). Confirming one is what is refused.
		return nil
	}

	engine, err := engineFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return err
	}

	for index, line := range lines {
		id, err := model.NewId()
		if err != nil {
			return err
		}
		fields := dmodel.DynamicFields{
			models.SalesOrderLineFieldId:               string(*id),
			models.SalesOrderLineFieldSalesOrderId:     orderId,
			models.SalesOrderLineFieldLineNumber:       int32(index + 1),
			models.SalesOrderLineFieldLineType:         string(models.SalesOrderLineTypeProduct),
			models.SalesOrderLineFieldProductVariantId: line.ProductVariantId,
			models.SalesOrderLineFieldUomId:            line.UomId,
			models.SalesOrderLineFieldOrderedQuantity:  line.Quantity,

			// True until SALES-048's product port can say otherwise. See the column's own note on
			// why the default leans this way.
			models.SalesOrderLineFieldRequiresFulfillment: true,

			models.SalesOrderLineFieldFulfilledQuantity: decimal.Zero,
			models.SalesOrderLineFieldReturnedQuantity:  decimal.Zero,

			models.SalesOrderLineFieldBaseUnitPrice:      line.UnitPrice,
			models.SalesOrderLineFieldEffectiveUnitPrice: line.UnitPrice,

			// The money columns are placeholders until the reprice that immediately follows. They
			// are written because the columns are NOT NULL, not because these are the answer.
			models.SalesOrderLineFieldGrossAmount:     decimal.Zero,
			models.SalesOrderLineFieldDiscountAmount:  decimal.Zero,
			models.SalesOrderLineFieldNetAmount:       decimal.Zero,
			models.SalesOrderLineFieldTaxRateSnapshot: decimal.Zero,
			models.SalesOrderLineFieldTaxAmount:       decimal.Zero,
			models.SalesOrderLineFieldFinalAmount:     decimal.Zero,

			models.SalesOrderLineFieldPricingSource: string(models.SalesOrderPricingSourceCatalogue),

			basemodel.FieldOrgId: orgId,
		}
		if line.ProductCode != "" {
			fields[models.SalesOrderLineFieldProductCodeSnapshot] = line.ProductCode
		}
		if line.ProductName != "" {
			fields[models.SalesOrderLineFieldProductNameSnapshot] = line.ProductName
		}
		if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
			return err
		}
	}
	return nil
}
