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

// Creating a sales order. Two rules here are load-bearing and neither is obvious from the fields.
//
// The channel is DERIVED from the sales point, never accepted: otherwise a kiosk could claim the
// `vdmc` channel while passing an eCommerce sales point id, booking itself into whichever channel
// priced better. A supplied code that disagrees is rejected outright.
//
// A duplicate idempotency key returns the ORIGINAL order, successfully - not a conflict. A retry
// must not create two orders, which is only true if the retry gets a success back: a gateway told
// "409" retries forever.

type CreateOrderParams struct {
	// SalesChannelCode is optional and is only ever CHECKED, never used to decide the channel.
	SalesChannelCode string

	SalesPointId string

	CustomerReference string
	CurrencyCode      string

	Lines []CreateOrderLine

	ExternalReference string

	// IdempotencyKey is unique per channel when present; absent means the caller accepts that a retry
	// creates a second order.
	IdempotencyKey string

	OrgId string
}

type CreateOrderLine struct {
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal

	// UnitPrice is the fallback when no pricelist item matches.
	UnitPrice decimal.Decimal

	ProductCode string
	ProductName string
}

type CreateOrderResult struct {
	SalesOrderId   string
	OrderNumber    string
	SalesChannelId string

	Pricing *RepriceResult

	// AlreadyExisted marks the idempotent replay path: success is returned either way, this says
	// whether anything was written.
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

func CreateOrder(
	ctx corectx.Context,
	params CreateOrderParams,
	taxSvc itExt.TaxCalculationExtService,
	products itExt.ProductVariantExtService,
	basisSvc itExt.ProductPricingBasisExtService,
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

	// Nothing withdrawn from sale may be ordered. Checked BEFORE the idempotency replay and any write:
	// a row written and then rejected would still be one a till could read and charge against.
	sellableErrs, err := assertVariantsSellable(ctx, params.Lines, products)
	if err != nil {
		return nil, nil, err
	}
	if sellableErrs != nil {
		return nil, sellableErrs, nil
	}

	// The idempotent replay, checked BEFORE writing anything. The unique index makes it safe under a
	// race: two simultaneous retries cannot both pass, and the loser's insert fails on the index.
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
		// A collision on the idempotency index means a concurrent retry won the race: the success
		// path, not a failure.
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

	// A create that could not price is a create that failed.
	priced, vErrs, err := RepriceOrder(ctx, orderId, taxSvc, policy, basisSvc)
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

// resolveSellingPlace loads the sales point first because it decides the channel; loading the
// channel from the request and checking the point against it would run the trust the wrong way.
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
		// A point whose channel is gone cannot sell. Reported against the point, the record the
		// caller named and the one an administrator would fix.
		return nil, nil, refuse("sales_point_id", ReasonChannelNotSellable,
			"this sales point references a sales channel that no longer exists"), nil
	}

	// A supplied code is CHECKED against the derived channel, never used instead of it.
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

// canSell requires both gates: archived is retired for good, suspended is stopped for now.
func canSell(record dmodel.DynamicFields, statusField, activeStatus string) bool {
	if boolOf(record, basemodel.FieldIsArchived) {
		return false
	}
	return stringOf(record, statusField) == activeStatus
}

// assertLinesRequestable checks what can be checked without reading another module. Variant
// existence needs inventory's product port, which Sales does not yet bind, so a line naming a
// nonexistent variant is stored and fails later.
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

// assertVariantsSellable refuses a variant Inventory has withdrawn. Separate from the pure
// assertLinesRequestable, which would otherwise be untestable without a container.
//
// A nil port PERMITS rather than refuses: it means a deployment with no master to be withdrawn
// from. The opposite reading from the tax port, which fails CLOSED because an unresolved tax
// silently undercharges the business.
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

	// Reported per LINE rather than per variant, because fixing the order means editing a line.
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

// writeDraftOrder uses one transaction because an order whose lines failed to write is an empty
// order - a legitimate state, so nothing downstream would flag it.
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

			// Zeroed rather than omitted: the reprice overwrites them, and a NOT NULL column with no
			// value would fail the insert.
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

func writeOrderLines(
	ctx corectx.Context, orderId, orgId string, lines []CreateOrderLine,
) error {
	if len(lines) == 0 {
		// An order with zero lines is a valid draft. Confirming one is what is refused.
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

			// True until the product port can say otherwise. See the column's own note.
			models.SalesOrderLineFieldRequiresFulfillment: true,

			models.SalesOrderLineFieldFulfilledQuantity: decimal.Zero,
			models.SalesOrderLineFieldReturnedQuantity:  decimal.Zero,

			models.SalesOrderLineFieldBaseUnitPrice:      line.UnitPrice,
			models.SalesOrderLineFieldEffectiveUnitPrice: line.UnitPrice,

			// Placeholders until the reprice that follows, written because the columns are NOT NULL.
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
