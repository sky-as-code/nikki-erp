package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The two read routes of the fulfillment feature.
//
// Both are engine actions rather than hand-written routes because Sales has no hand-written REST
// layer at all — every resource is served by its engine — and adding one for two read endpoints
// would introduce a second routing mechanism to maintain. The CR's paths differ in spelling only;
// the payloads and semantics are the ones it specifies.
//
// Both answer `read`. Neither changes anything, and requiring a write permission to look at what a
// customer is owed would put the question behind a role that can also alter the answer.

const (
	ActionFulfillments       = "fulfillments"
	ActionFulfillmentTargets = "fulfillment_targets"
	paramItems               = "items"
	paramProductVariantId    = "product_variant_id"
	paramQuantity            = "quantity"
	paramOrgId               = "org_id"
	reasonItemsMalformed     = "sales_order_fulfillment.items_malformed"
	reasonQuantityMalformed  = "sales_order_fulfillment.quantity_malformed"
)

// defineSalesOrderFulfillmentViewActions rides on the sales_order engine rather than the fulfillment
// one: both questions are asked about an ORDER, and hanging them off the fulfillment resource would
// force a caller to know a fulfillment id before it could find out whether any exist.
func defineSalesOrderFulfillmentViewActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionFulfillments,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/fulfillments",
			Permission:  drif.PermissionRead,
			MainProcess: processOrderFulfillments,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionFulfillmentTargets,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ActionFulfillmentTargets + "/search",
			Permission:  drif.PermissionRead,
			MainProcess: processFulfillmentTargetSearch,
		}),
	)
}

// processOrderFulfillments answers what an order's deliveries owe. Refund state is deliberately
// absent from fulfillment_status; the quantities carry it instead, and pending_refund_qty is what
// separates "still owed" from "may be attempted now".
func processOrderFulfillments(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	views, err := services.ViewOrderFulfillments(ctx, readStringParam(input.Params, paramId))
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

	return &drif.ActionResult{
		HasData: true,
		Data:    map[string]any{"fulfillments": fulfillments},
	}, nil
}

// processFulfillmentTargetSearch shortlists the kiosks that could supply a basket.
//
// The response says advisory in as many words. It takes no lock, so a target reported able to supply
// may be emptied by another sale before the customer picks it; only reserving secures anything, and
// a client that treated this as a guarantee would promise goods it cannot deliver.
func processFulfillmentTargetSearch(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	items, vErrs := readAvailabilityItems(input.Params)
	if vErrs != nil {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	options, err := services.SearchFulfillmentTargets(
		ctx, readStringParam(input.Params, paramOrgId), items, fulfillmentReservations)
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

	return &drif.ActionResult{
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

// readDecimalValue accepts every numeric shape a JSON body can arrive in: a whole number comes back
// from the decoder as a float64, and a client sending an exact decimal sends a string.
func readDecimalValue(value any) (decimal.Decimal, bool) {
	switch typed := value.(type) {
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			return decimal.Zero, false
		}
		return parsed, true
	case float64:
		return decimal.NewFromFloat(typed), true
	case int:
		return decimal.NewFromInt(int64(typed)), true
	case int64:
		return decimal.NewFromInt(typed), true
	}
	return decimal.Zero, false
}

func itemsViolation(key, message string) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.SalesOrderFulfillmentSchemaName, key, message))
	return vErrs
}
