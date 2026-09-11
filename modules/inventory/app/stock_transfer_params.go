package app

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

const (
	paramReturnLines        = "lines"
	paramReturnMoveId       = "move_id"
	paramReturnQuantity     = "quantity"
	paramItemSourceItemId   = "source_item_id"
	paramItemVariantId      = "product_variant_id"
	paramItemQuantity       = "quantity"
	paramItemUomId          = "uom_id"
	paramItemFailureCode    = "failure_code"
	paramItemFailureMessage = "failure_message"
)

// readReturnRequest reads the optional per-line quantities from the request body. Absent `lines`
// means "return everything still returnable"; a caller wanting a partial return names the moves.
func readReturnRequest(params dmodel.DynamicFields) (services.ReturnRequest, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	request := services.ReturnRequest{}

	raw, present := params[paramReturnLines]
	if !present || raw == nil {
		return request, vErrs
	}
	items, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName, "stock_return.lines_malformed",
			"'lines' must be a list of {move_id, quantity} entries"))
		return request, vErrs
	}

	for _, item := range items {
		line, vErr := readReturnLine(item)
		if vErr != nil {
			vErrs.Append(*vErr)
			continue
		}
		request.Lines = append(request.Lines, line)
	}
	return request, vErrs
}

func readReturnLine(item any) (services.ReturnLineRequest, *ft.ClientErrorItem) {
	fields, ok := item.(map[string]any)
	if !ok {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_malformed",
			"each return line must be an object with move_id and quantity")
	}
	moveId, _ := fields[paramReturnMoveId].(string)
	if moveId == "" {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_move_id_required",
			"each return line must name a move_id")
	}
	quantity, vErr := readItemDecimal(fields, paramReturnQuantity)
	if vErr != nil {
		return services.ReturnLineRequest{}, ft.NewBusinessViolation(
			models.StockTransferSchemaName, "stock_return.line_quantity_malformed",
			"return line for move '"+moveId+"' must carry a decimal quantity")
	}
	return services.ReturnLineRequest{MoveId: moveId, Quantity: *quantity}, nil
}

// readSourceItems reads the lines of a hold request.
func readSourceItems(params dmodel.DynamicFields, field string) ([]itStock.SourceReservationItem, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	entries, ok := readEntries(params, field, vErrs)
	if !ok {
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
func readResultItems(params dmodel.DynamicFields, field string) ([]itStock.FulfillmentResultItem, *ft.ClientErrors) {
	vErrs := ft.NewClientErrors()
	entries, ok := readEntries(params, field, vErrs)
	if !ok {
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

// readEntries reads a list param. Absence is legitimate (nothing to act on); a present value that
// is not a list is a violation.
func readEntries(params dmodel.DynamicFields, field string, vErrs *ft.ClientErrors) ([]any, bool) {
	raw, present := params[field]
	if !present || raw == nil {
		return nil, false
	}
	entries, ok := raw.([]any)
	if !ok {
		vErrs.Append(*ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.items_malformed",
			"'"+field+"' must be a list of {source_item_id, product_variant_id, quantity} entries"))
		return nil, false
	}
	return entries, true
}

// readItemDecimal reads one item's quantity, accepting the string form as well as a bare number.
// A malformed value is a violation rather than a silent zero, which would reserve nothing and
// look like success.
func readItemDecimal(fields map[string]any, name string) (*decimal.Decimal, *ft.ClientErrorItem) {
	raw, present := fields[name]
	if !present || raw == nil {
		return nil, ft.NewBusinessViolation(models.StockTransferSchemaName,
			"stock_transfer.quantity_required", "each item must name a '"+name+"'")
	}
	value, ok := toDecimal(raw)
	if !ok {
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
