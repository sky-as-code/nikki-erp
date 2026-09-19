package services

import (
	"github.com/shopspring/decimal"
	"time"

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

	// Fulfillment is what the client asked for about delivery. Both parts are optional and neither
	// is trusted: they are recorded as a request and settled at confirm, where the policy is
	// resolved and validated. A client can never send the policy snapshot fields themselves — those
	// are copied from the resolved method, and accepting them would let a caller choose its own
	// refund rules.
	Fulfillment CreateOrderFulfillment

	// EstimatedTotalPrice is the order total the client worked out for itself. Recorded for
	// reconciliation and never charged; nil means the client did not send one.
	EstimatedTotalPrice *decimal.Decimal

	// ValidUntil is the payment deadline, settable only at create and never extended. Nil means
	// the order does not expire.
	ValidUntil *time.Time

	OrgId string

	// The channel's automation flags, read by CreateOrder from the channel it resolved and copied
	// onto the draft. Unexported: a caller cannot choose its own automation.
	autoConfirmOrder  bool
	autoConfirmRefund bool
}

// CreateOrderFulfillment is the client's delivery request. Empty means "use the defaults", which
// is what an ordinary sale sends.
type CreateOrderFulfillment struct {
	// FulfillmentMethodId names a method outright, overriding the point and channel defaults.
	FulfillmentMethodId string

	// TargetOutletId names where to collect. Required before confirm only for a method whose
	// strategy is customer_selected_outlet; supplying it for any other method is harmless and
	// simply takes precedence over the derived target.
	TargetOutletId string
}

type CreateOrderLine struct {
	ProductVariantId string
	UomId            string
	Quantity         decimal.Decimal

	Allocations []CreateOrderLineAllocation

	// UnitPrice is the fallback when no pricelist item matches.
	UnitPrice decimal.Decimal

	// EstimatedPrice is the unit price the client worked out for itself, recorded so a POS or kiosk
	// showing a different number can be reconciled against what Sales charged. Never priced from:
	// nil simply means the client did not send one.
	EstimatedPrice *decimal.Decimal

	ProductCode string
	ProductName string
}

type CreateOrderLineAllocation struct {
	LocationId string
	Quantity   decimal.Decimal
}

type CreateOrderResult struct {
	SalesOrderId   string
	OrderNumber    string
	SalesChannelId string

	Pricing *RepriceResult

	// AlreadyExisted marks the idempotent replay path: success is returned either way, this says
	// whether anything was written.
	AlreadyExisted bool

	// AutoConfirmOrder is the channel's setting as snapshotted onto the draft. The caller confirms
	// the draft it just got back when this is set; create itself never inserts a confirmed order.
	AutoConfirmOrder bool
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
	params.autoConfirmOrder = boolOf(channel, models.SalesChannelFieldAutoConfirmOrder)
	params.autoConfirmRefund = boolOf(channel, models.SalesChannelFieldAutoConfirmRefund)

	if vErrs := assertLinesRequestable(params.Lines); vErrs != nil {
		return nil, vErrs, nil
	}
	if vErrs := assertValidUntilInFuture(params.ValidUntil, time.Now().UTC()); vErrs != nil {
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

	// Reported, never enforced: the order stands at the price Sales calculated whatever the client
	// thought it would be.
	err = RecordEstimatedPriceDivergence(ctx, EstimatedPriceCheckParams{
		SalesOrderId: orderId,
		OrgId:        params.OrgId,
		Estimated:    params.EstimatedTotalPrice,
		Actual:       priced.GrandTotal,
		Tolerance:    policy.EstimatedPriceTolerance,
	})
	if err != nil {
		return nil, nil, err
	}

	return &CreateOrderResult{
		SalesOrderId:     orderId,
		OrderNumber:      orderNumber,
		SalesChannelId:   channelId,
		Pricing:          priced,
		AutoConfirmOrder: params.autoConfirmOrder,
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
	engineRepo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.SalesOrderFieldSalesChannelId, dmodel.Equals, channelId))
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.SalesOrderFieldIdempotencyKey, dmodel.Equals, key))

	found, err := engineRepo.Search(ctx, dyn.RepoSearchParam{
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
		orderRepo, err := repoFor(models.SalesOrderSchemaName)
		if err != nil {
			return err
		}

		fields := dmodel.DynamicFields{
			models.SalesOrderFieldId:             orderId,
			models.SalesOrderFieldOrderNumber:    orderNumber,
			models.SalesOrderFieldSalesChannelId: channelId,
			models.SalesOrderFieldSalesPointId:   params.SalesPointId,
			models.SalesOrderFieldCurrencyCode:   params.CurrencyCode,

			models.SalesOrderFieldStatus: string(models.SalesOrderStatusDraft),

			// Entering draft IS entering a stage, so the count starts at one here rather than at the
			// first move away from it.
			models.SalesOrderFieldStageVersion: int32(1),

			models.SalesOrderFieldPaymentStatus:     string(models.SalesOrderPaymentStatusUnpaid),
			models.SalesOrderFieldFulfillmentStatus: string(models.SalesOrderFulfillmentStatusPending),
			models.SalesOrderFieldInvoiceStatus:     string(models.SalesOrderInvoiceStatusNotRequested),

			// Calculated from the request's own authentication, never from params: a client able to
			// assert it was choosing its own refund policy. See DeriveCustomerIdentityMode.
			models.SalesOrderFieldCustomerIdentityMode: string(DeriveCustomerIdentityMode(ctx)),

			// Zeroed rather than omitted: the reprice overwrites them, and a NOT NULL column with no
			// value would fail the insert.
			models.SalesOrderFieldSubtotal:      decimal.Zero,
			models.SalesOrderFieldDiscountTotal: decimal.Zero,
			models.SalesOrderFieldTaxTotal:      decimal.Zero,
			models.SalesOrderFieldGrandTotal:    decimal.Zero,

			basemodel.FieldOrgId: orgId,
			// Direct repository inserts do not apply the archivable model's default.
			basemodel.FieldIsArchived: false,
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
		if params.ValidUntil != nil {
			fields[models.SalesOrderFieldValidUntil] = model.ModelDateTime(params.ValidUntil.UTC())
		}

		// The channel's two automation flags are copied here, once: every later decision about
		// this order reads the copy, so the channel changing its mind never reaches it.
		fields[models.SalesOrderFieldAutoConfirmOrder] = params.autoConfirmOrder
		fields[models.SalesOrderFieldAutoConfirmRefund] = params.autoConfirmRefund
		if params.EstimatedTotalPrice != nil {
			fields[models.SalesOrderFieldEstimatedTotalPrice] = *params.EstimatedTotalPrice
		}
		// Recorded as asked for, not validated here: the method is resolved at confirm, where the
		// channel, the target's readiness and the customer's identity are all known together.
		if params.Fulfillment.FulfillmentMethodId != "" {
			fields[models.SalesOrderFieldRequestedFulfillmentMethodId] = params.Fulfillment.FulfillmentMethodId
		}
		if params.Fulfillment.TargetOutletId != "" {
			fields[models.SalesOrderFieldRequestedTargetOutletId] = params.Fulfillment.TargetOutletId
		}

		if _, err := orderRepo.Insert(tranxCtx, fields); err != nil {
			return err
		}

		// null -> draft. An order entering its first stage announces itself like any other stage
		// change, so a consumer tracking the lifecycle sees the sale from its beginning rather than
		// first hearing of it at confirmation. Written in the creating transaction, so an order that
		// failed to insert announces nothing.
		if err := RecordOrderStageChanged(tranxCtx, OrderStageChangedParams{
			Order:         fields,
			PreviousStage: "",
			CurrentStage:  string(models.SalesOrderStatusDraft),
			StageVersion:  1,
		}); err != nil {
			return err
		}
		lineIds, err := writeOrderLines(tranxCtx, orderId, orgId, params.Lines)
		if err != nil {
			return err
		}
		return writeOrderLineAllocations(tranxCtx, lineIds, orgId, params.Lines)
	})
	if err != nil {
		return "", "", err
	}
	return orderId, orderNumber, nil
}

func writeOrderLines(
	ctx corectx.Context, orderId, orgId string, lines []CreateOrderLine,
) ([]string, error) {
	if len(lines) == 0 {
		// An order with zero lines is a valid draft. Confirming one is what is refused.
		return nil, nil
	}

	engineRepo, err := repoFor(models.SalesOrderLineSchemaName)
	if err != nil {
		return nil, err
	}

	lineIds := make([]string, 0, len(lines))
	for index, line := range lines {
		id, err := model.NewId()
		if err != nil {
			return nil, err
		}
		lineId := string(*id)
		fields := dmodel.DynamicFields{
			models.SalesOrderLineFieldId:               lineId,
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

			basemodel.FieldOrgId:      orgId,
			basemodel.FieldIsArchived: false,
		}
		if line.ProductCode != "" {
			fields[models.SalesOrderLineFieldProductCodeSnapshot] = line.ProductCode
		}
		if line.ProductName != "" {
			fields[models.SalesOrderLineFieldProductNameSnapshot] = line.ProductName
		}
		if line.EstimatedPrice != nil {
			fields[models.SalesOrderLineFieldEstimatedPrice] = *line.EstimatedPrice
		}
		if _, err := engineRepo.Insert(ctx, fields); err != nil {
			return nil, err
		}
		lineIds = append(lineIds, lineId)
	}
	return lineIds, nil
}

// writeOrderLineAllocations persists the requested stock split after its parent lines exist.
// lineIds and lines intentionally share an index: writeOrderLines creates the line at index i
// before this function writes the allocations of input line i, so an allocation never has to
// rediscover its parent by variant (which would be ambiguous for duplicate variants).
func writeOrderLineAllocations(
	ctx corectx.Context, lineIds []string, orgId string, lines []CreateOrderLine,
) error {
	hasAllocations := false
	for _, line := range lines {
		if len(line.Allocations) > 0 {
			hasAllocations = true
			break
		}
	}
	if !hasAllocations {
		return nil
	}

	engineRepo, err := repoFor(models.SalesOrderLineAllocationSchemaName)
	if err != nil {
		return err
	}

	for lineIndex, line := range lines {
		for _, allocation := range line.Allocations {
			id, err := model.NewId()
			if err != nil {
				return err
			}

			fields := dmodel.DynamicFields{
				models.SalesOrderLineAllocationFieldId:               string(*id),
				models.SalesOrderLineAllocationFieldSalesOrderLineId: lineIds[lineIndex],
				models.SalesOrderLineAllocationFieldSourceLocationId: allocation.LocationId,
				models.SalesOrderLineAllocationFieldQuantity:         allocation.Quantity,
				basemodel.FieldOrgId:                                 orgId,
				basemodel.FieldIsArchived:                            false,
			}
			if _, err := engineRepo.Insert(ctx, fields); err != nil {
				return err
			}
		}
	}
	return nil
}
