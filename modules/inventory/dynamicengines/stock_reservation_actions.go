package dynamicengines

import (
	stdErr "errors"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The reservation operations addressed by the DEMAND rather than by a transfer id.
//
// Every other movement action takes ':id' because the caller is sequencing a document it created and
// holds the id for. These four are for a caller in another module — or in another process — that
// knows only what it asked for: which fulfillment, which line. Making it look the transfer up first
// would mean publishing a query whose only purpose is to find an id it should not need to hold, and
// would leave a window in which the id it found is no longer the current hold.
//
// They are collection-level POSTs for that reason: the demand names the subject, not the URL.
const (
	// Holding and releasing stock are the same powers the id-addressed actions carry, so they reuse
	// those permissions rather than inventing parallel ones: a role able to reserve a transfer can
	// reserve for a demand, and splitting them would let one be granted without the other by
	// accident.
	ActionReserveForSource       = "reserve_for_source"
	ActionReleaseForSource       = "release_for_source"
	ActionReallocateReservation  = "reallocate_reservation"
	ActionApplyFulfillmentResult = "apply_fulfillment_result"

	// Recording what physically happened is its own power, distinct from validate: validate is an
	// operator saying "ship this document", while this is a machine reporting what it managed to
	// hand over. The caller is a service, and granting it the ability to validate arbitrary
	// documents would be far more than it needs.
	PermissionApplyFulfillmentResult = "apply_fulfillment_result"
)

const (
	paramSourceType         = "source_type"
	paramSourceId           = "source_id"
	paramSourceItems        = "items"
	paramHoldLocationId     = "location_id"
	paramToLocationId       = "to_location_id"
	paramOperationTypeId    = "operation_type_id"
	paramOriginReference    = "origin_reference"
	paramEventId            = "event_id"
	paramExecutorLocationId = "executor_location_id"
	paramSuccessfulItems    = "successful_items"
	paramFailedItems        = "failed_items"
	paramItemSourceItemId   = "source_item_id"
	paramItemVariantId      = "product_variant_id"
	paramItemQuantity       = "quantity"
	paramItemUomId          = "uom_id"
	paramItemFailureCode    = "failure_code"
	paramItemFailureMessage = "failure_message"
)

func defineStockReservationSourceActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReserveForSource,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ActionReserveForSource,
			Permission:  PermissionReserve,
			MainProcess: processReserveForSource,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReleaseForSource,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ActionReleaseForSource,
			Permission:  PermissionUnreserve,
			MainProcess: processReleaseForSource,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionReallocateReservation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ActionReallocateReservation,
			Permission:  PermissionReserve,
			MainProcess: processReallocateReservation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionApplyFulfillmentResult,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ActionApplyFulfillmentResult,
			Permission:  PermissionApplyFulfillmentResult,
			MainProcess: processApplyFulfillmentResult,
		}),
	)
}

func processReserveForSource(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}

	items, vErrs := readSourceItems(input.Params, paramSourceItems)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := service.ReserveForSource(ctx, itStock.SourceReservationRequest{
		SourceType:      readStringField(input.Params, paramSourceType),
		SourceId:        readStringField(input.Params, paramSourceId),
		OrgId:           readStringField(input.Params, models.StockTransferFieldOrgId),
		LocationId:      readStringField(input.Params, paramHoldLocationId),
		OperationTypeId: readStringField(input.Params, paramOperationTypeId),
		OriginReference: readStringField(input.Params, paramOriginReference),
		Items:           items,
	})
	return toSourceReservationResult(result, err)
}

func processReleaseForSource(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.ReleaseReservationBySource(ctx,
		readStringField(input.Params, paramSourceType),
		readStringField(input.Params, paramSourceId))
	return toMutateActionResult(result, err)
}

func processReallocateReservation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}

	items, vErrs := readSourceItems(input.Params, paramSourceItems)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := service.ReallocateReservation(ctx, itStock.ReservationReallocationRequest{
		SourceType:      readStringField(input.Params, paramSourceType),
		SourceId:        readStringField(input.Params, paramSourceId),
		OrgId:           readStringField(input.Params, models.StockTransferFieldOrgId),
		ToLocationId:    readStringField(input.Params, paramToLocationId),
		OperationTypeId: readStringField(input.Params, paramOperationTypeId),
		OriginReference: readStringField(input.Params, paramOriginReference),
		Items:           items,
	})
	return toSourceReservationResult(result, err)
}

func processApplyFulfillmentResult(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := transferServiceOf(input)
	if err != nil {
		return nil, err
	}

	successful, vErrs := readResultItems(input.Params, paramSuccessfulItems)
	failed, failedErrs := readResultItems(input.Params, paramFailedItems)
	vErrs.ConcatPtr(failedErrs)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := service.ApplyFulfillmentResult(ctx, itStock.FulfillmentResultRequest{
		EventId:            readStringField(input.Params, paramEventId),
		SourceType:         readStringField(input.Params, paramSourceType),
		SourceId:           readStringField(input.Params, paramSourceId),
		OrgId:              readStringField(input.Params, models.StockTransferFieldOrgId),
		ExecutorLocationId: readStringField(input.Params, paramExecutorLocationId),
		SuccessfulItems:    successful,
		FailedItems:        failed,
	})
	if err != nil {
		return nil, err
	}
	return &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.ClientErrors.Count() == 0,
		Data:         result,
	}, nil
}

// toSourceReservationResult shapes a hold's outcome for the REST layer. A refusal travels as
// ClientErrors with no Go error, which is what makes the answer 400 rather than 500: a location
// that cannot supply the goods is something the caller fixes by choosing another.
func toSourceReservationResult(
	result *itStock.SourceReservationResult, err error,
) (*drif.ActionResult, error) {
	if err != nil {
		return nil, err
	}
	return &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.ClientErrors.Count() == 0,
		Data:         result,
	}, nil
}

// readSourceItems reads the lines of a hold request.
func readSourceItems(
	params dmodel.DynamicFields, field string,
) ([]itStock.SourceReservationItem, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()

	raw, present := params[field]
	if !present || raw == nil {
		return nil, vErrs
	}
	entries, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.items_malformed",
			"'"+field+"' must be a list of {source_item_id, product_variant_id, quantity} entries"))
		return nil, vErrs
	}

	items := make([]itStock.SourceReservationItem, 0, len(entries))
	for _, entry := range entries {
		fields, ok := entry.(map[string]any)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName,
				"stock_transfer.items_malformed", "each entry of '"+field+"' must be an object"))
			continue
		}
		quantity, vErr := readItemDecimal(fields, paramItemQuantity)
		if vErr != nil {
			vErrs.Append(*vErr)
			continue
		}
		items = append(items, itStock.SourceReservationItem{
			SourceItemId:     readItemString(fields, paramItemSourceItemId),
			ProductVariantId: readItemString(fields, paramItemVariantId),
			UomId:            readItemString(fields, paramItemUomId),
			Quantity:         *quantity,
		})
	}
	return items, vErrs
}

// readResultItems reads one side of a dispense result.
func readResultItems(
	params dmodel.DynamicFields, field string,
) ([]itStock.FulfillmentResultItem, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()

	raw, present := params[field]
	if !present || raw == nil {
		return nil, vErrs
	}
	entries, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.items_malformed",
			"'"+field+"' must be a list of {source_item_id, product_variant_id, quantity} entries"))
		return nil, vErrs
	}

	items := make([]itStock.FulfillmentResultItem, 0, len(entries))
	for _, entry := range entries {
		fields, ok := entry.(map[string]any)
		if !ok {
			vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName,
				"stock_transfer.items_malformed", "each entry of '"+field+"' must be an object"))
			continue
		}
		quantity, vErr := readItemDecimal(fields, paramItemQuantity)
		if vErr != nil {
			vErrs.Append(*vErr)
			continue
		}
		items = append(items, itStock.FulfillmentResultItem{
			SourceItemId:     readItemString(fields, paramItemSourceItemId),
			ProductVariantId: readItemString(fields, paramItemVariantId),
			Quantity:         *quantity,
			FailureCode:      readItemString(fields, paramItemFailureCode),
			FailureMessage:   readItemString(fields, paramItemFailureMessage),
		})
	}
	return items, vErrs
}

// readItemDecimal reads one item's quantity, accepting the string form JSON carries money and
// quantities in
// as well as a bare number. A malformed value is a violation rather than a silent zero, which would
// reserve nothing and look like success.
func readItemDecimal(fields map[string]any, name string) (*decimal.Decimal, *ft.ClientErrorItem) {
	raw, present := fields[name]
	if !present || raw == nil {
		return nil, ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.quantity_required", "each item must name a '"+name+"'")
	}

	var value decimal.Decimal
	switch typed := raw.(type) {
	case string:
		parsed, err := decimal.NewFromString(typed)
		if err != nil {
			return nil, ft.NewBusinessViolation(models.StockTransferSchemaName,
				"stock_transfer.quantity_malformed", "'"+name+"' is not a number: "+typed)
		}
		value = parsed
	case float64:
		value = decimal.NewFromFloat(typed)
	case int:
		value = decimal.NewFromInt(int64(typed))
	case int64:
		value = decimal.NewFromInt(typed)
	default:
		return nil, ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.quantity_malformed", "'"+name+"' is not a number")
	}
	return &value, nil
}

func readItemString(fields map[string]any, name string) string {
	raw, present := fields[name]
	if !present || raw == nil {
		return ""
	}
	if typed, ok := raw.(string); ok {
		return typed
	}
	return ""
}
