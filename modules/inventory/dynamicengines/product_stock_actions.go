package dynamicengines

import (
	stdErr "errors"
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The reads that let the Product UI show stock it does not own. They hang off the quant engine
// because what they read is quants, and they are all reads: none creates a movement, reserves
// anything or changes a balance.
//
// The batch action exists so a listing summarises a whole page in one request; calling the
// single-variant action once per row would be an N+1. The single form is for a detail page.

const (
	ActionVariantStockSummary   = "variant_stock_summary"
	ActionVariantStockSummaries = "variant_stock_summaries"
	ActionTemplateStockSummary  = "template_stock_summary"
	ActionStockByWarehouse      = "stock_by_warehouse"
	ActionStockByLocation       = "stock_by_location"
	ActionProductUsage          = "product_usage"
)

// One permission covers all of them: they are the same power sliced by which product you look at,
// and granting them separately would allow "may see a total but not the rows behind it".
const PermissionReadProductStock = "read_product_stock"

const (
	paramSummaryVariantId  = "product_variant_id"
	paramSummaryVariantIds = "product_variant_ids"
	paramSummaryTemplateId = "product_template_id"
	paramSummaryWarehouse  = "warehouse_id"
)

// variantSummaryResponse is the wire shape of one variant's stock. Quantities are strings, not
// numbers: they are decimals, and a JSON number is float64 at the far end, which is how a quantity
// acquires a rounding error it never had in the database.
type variantSummaryResponse struct {
	ProductVariantId string `json:"productVariantId,omitempty"`
	OnHand           string `json:"onHand"`
	Reserved         string `json:"reserved"`
	Available        string `json:"available"`
	Forecasted       string `json:"forecasted"`
	InTransit        string `json:"inTransit"`
	LocationCount    int    `json:"locationCount"`
	WarehouseCount   int    `json:"warehouseCount"`
	BaseUomId        string `json:"baseUomId,omitempty"`
	LastMovementAt   string `json:"lastMovementAt,omitempty"`

	// Truncated warns that the totals are partial, so a UI can say so instead of presenting an
	// incomplete number as the whole one.
	Truncated bool `json:"truncated,omitempty"`
}

type templateSummaryResponse struct {
	Summary  variantSummaryResponse   `json:"summary"`
	Variants []templateVariantRowJson `json:"variants"`
}

type templateVariantRowJson struct {
	ProductVariantId string `json:"productVariantId"`
	Sku              string `json:"sku,omitempty"`
	CombinationKey   string `json:"combinationKey,omitempty"`

	OnHand     string `json:"onHand"`
	Reserved   string `json:"reserved"`
	Available  string `json:"available"`
	Forecasted string `json:"forecasted"`
	InTransit  string `json:"inTransit"`
}

type warehouseStockRowJson struct {
	WarehouseId     string `json:"warehouseId,omitempty"`
	WarehouseCode   string `json:"warehouseCode,omitempty"`
	WarehouseName   string `json:"warehouseName,omitempty"`
	WarehouseStatus string `json:"warehouseStatus,omitempty"`

	OnHand    string `json:"onHand"`
	Reserved  string `json:"reserved"`
	Available string `json:"available"`
}

type locationStockRowJson struct {
	LocationId     string `json:"locationId"`
	LocationCode   string `json:"locationCode,omitempty"`
	LocationName   string `json:"locationName,omitempty"`
	LocationStatus string `json:"locationStatus,omitempty"`
	WarehouseId    string `json:"warehouseId,omitempty"`

	OnHand    string `json:"onHand"`
	Reserved  string `json:"reserved"`
	Available string `json:"available"`
}

type productUsageResponse struct {
	OnHandQuantity    string `json:"onHandQuantity"`
	ReservedQuantity  string `json:"reservedQuantity"`
	OpenMoveCount     int    `json:"openMoveCount"`
	OpenTransferCount int    `json:"openTransferCount"`

	// CanArchive is the reader's own verdict on its four numbers, so a caller does not restate the
	// rule and risk restating it differently.
	CanArchive bool `json:"canArchive"`
}

// defineProductStockActions adds the product-facing reads to the quant engine.
func defineProductStockActions(engine drif.DynamicResourceEngine) error {
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionVariantStockSummary,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "variant_stock_summary",
			Permission:  PermissionReadProductStock,
			MainProcess: processVariantStockSummary,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionVariantStockSummaries,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "variant_stock_summaries",
			Permission:  PermissionReadProductStock,
			MainProcess: processVariantStockSummaries,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionTemplateStockSummary,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "template_stock_summary",
			Permission:  PermissionReadProductStock,
			MainProcess: processTemplateStockSummary,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionStockByWarehouse,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "stock_by_warehouse",
			Permission:  PermissionReadProductStock,
			MainProcess: processStockByWarehouse,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionStockByLocation,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "stock_by_location",
			Permission:  PermissionReadProductStock,
			MainProcess: processStockByLocation,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionProductUsage,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    "product_usage",
			Permission:  PermissionReadProductStock,
			MainProcess: processProductUsage,
		}),
	)
}

// productStockReaderOf reaches the quant service, which implements both product-facing ports. A
// failed assertion is a wiring bug rather than a bad request, so it is a plain error.
func productStockReaderOf(input drif.ProcessInput) (*services.StockQuantDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.StockQuantDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the stock quant engine is not running the derived quant service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func processVariantStockSummary(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	variantId := readStringField(input.Params, paramSummaryVariantId)
	result, err := service.GetVariantSummaries(
		ctx, itStock.GetVariantSummariesQuery{VariantIds: []string{variantId}})
	if err != nil {
		return nil, err
	}

	response := toVariantSummaryJson(variantId, result.Data.Summaries[variantId])
	return &drif.ActionResult{Data: response, HasData: true}, nil
}

// processVariantStockSummaries is the batch form, for a listing summarising its whole page.
func processVariantStockSummaries(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	variantIds := readStringSliceField(input.Params, paramSummaryVariantIds)
	result, err := service.GetVariantSummaries(
		ctx, itStock.GetVariantSummariesQuery{VariantIds: variantIds})
	if err != nil {
		return nil, err
	}

	// Keyed by variant id so a caller matches each summary to its row without relying on order,
	// which a map-backed result does not preserve anyway.
	response := make(map[string]variantSummaryResponse, len(result.Data.Summaries))
	for variantId, summary := range result.Data.Summaries {
		response[variantId] = toVariantSummaryJson("", summary)
	}
	return &drif.ActionResult{Data: response, HasData: true}, nil
}

func processTemplateStockSummary(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	result, err := service.GetTemplateSummary(ctx, itStock.GetTemplateSummaryQuery{
		TemplateId: readStringField(input.Params, paramSummaryTemplateId),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]templateVariantRowJson, 0, len(result.Data.Variants))
	for _, row := range result.Data.Variants {
		rows = append(rows, templateVariantRowJson{
			ProductVariantId: string(row.VariantId),
			Sku:              row.Sku,
			CombinationKey:   row.CombinationKey,
			OnHand:           row.Summary.OnHand.String(),
			Reserved:         row.Summary.Reserved.String(),
			Available:        row.Summary.Available.String(),
			Forecasted:       row.Summary.Forecasted.String(),
			InTransit:        row.Summary.InTransit.String(),
		})
	}

	return &drif.ActionResult{
		Data: templateSummaryResponse{
			Summary:  toVariantSummaryJson("", result.Data.Summary),
			Variants: rows,
		},
		HasData: true,
	}, nil
}

func processStockByWarehouse(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	result, err := service.GetStockByWarehouse(ctx, itStock.GetStockByWarehouseQuery{
		VariantId: readStringField(input.Params, paramSummaryVariantId),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]warehouseStockRowJson, 0, len(result.Data.Rows))
	for _, row := range result.Data.Rows {
		rows = append(rows, warehouseStockRowJson{
			WarehouseId:     idOrEmpty(row.WarehouseId),
			WarehouseCode:   row.WarehouseCode,
			WarehouseName:   row.WarehouseName,
			WarehouseStatus: row.WarehouseStatus,
			OnHand:          row.OnHand.String(),
			Reserved:        row.Reserved.String(),
			Available:       row.Available.String(),
		})
	}
	return &drif.ActionResult{Data: rows, HasData: true}, nil
}

func processStockByLocation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	result, err := service.GetStockByLocation(ctx, itStock.GetStockByLocationQuery{
		VariantId:   readStringField(input.Params, paramSummaryVariantId),
		WarehouseId: readStringField(input.Params, paramSummaryWarehouse),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]locationStockRowJson, 0, len(result.Data.Rows))
	for _, row := range result.Data.Rows {
		rows = append(rows, locationStockRowJson{
			LocationId:     string(row.LocationId),
			LocationCode:   row.LocationCode,
			LocationName:   row.LocationName,
			LocationStatus: row.LocationStatus,
			WarehouseId:    idOrEmpty(row.WarehouseId),
			OnHand:         row.OnHand.String(),
			Reserved:       row.Reserved.String(),
			Available:      row.Available.String(),
		})
	}
	return &drif.ActionResult{Data: rows, HasData: true}, nil
}

// processProductUsage reports what would be stranded if a variant were archived now, so a UI can
// explain a refusal before the user attempts it. Not the enforcement point: the archive is guarded
// independently in the product service, which still refuses a client that skipped this call.
func processProductUsage(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := productStockReaderOf(input)
	if err != nil {
		return nil, err
	}

	result, err := service.GetProductUsage(ctx, itStock.GetProductUsageQuery{
		VariantId: readStringField(input.Params, paramSummaryVariantId),
	})
	if err != nil {
		return nil, err
	}

	usage := result.Data.Usage
	return &drif.ActionResult{
		Data: productUsageResponse{
			OnHandQuantity:    usage.OnHandQuantity.String(),
			ReservedQuantity:  usage.ReservedQuantity.String(),
			OpenMoveCount:     usage.OpenMoveCount,
			OpenTransferCount: usage.OpenTransferCount,
			CanArchive:        usage.IsEmpty(),
		},
		HasData: true,
	}, nil
}

func toVariantSummaryJson(
	variantId string, summary itStock.VariantStockSummary,
) variantSummaryResponse {
	response := variantSummaryResponse{
		ProductVariantId: variantId,
		OnHand:           decimalOrZero(summary.OnHand),
		Reserved:         decimalOrZero(summary.Reserved),
		Available:        decimalOrZero(summary.Available),
		Forecasted:       decimalOrZero(summary.Forecasted),
		InTransit:        decimalOrZero(summary.InTransit),
		LocationCount:    summary.LocationCount,
		WarehouseCount:   summary.WarehouseCount,
		Truncated:        summary.Truncated,
	}
	if summary.BaseUomId != nil {
		response.BaseUomId = string(*summary.BaseUomId)
	}
	if summary.LastMovementAt != nil {
		response.LastMovementAt = summary.LastMovementAt.Format(time.RFC3339)
	}
	return response
}

// decimalOrZero renders a quantity, spelling zero as "0" rather than blank so a UI never has to
// decide what an empty string means.
func decimalOrZero(value decimal.Decimal) string {
	return value.String()
}

func idOrEmpty[T ~string](value *T) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// readStringSliceField reads a list of ids out of the action params. A decoded JSON array arrives
// as []any of strings; []string is accepted too for params built in Go. Anything else reads as
// absent rather than erroring — an unparseable list means there is nothing to summarise.
func readStringSliceField(params dmodel.DynamicFields, field string) []string {
	value, ok := params[field]
	if !ok || value == nil {
		return nil
	}

	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && text != "" {
				result = append(result, text)
			}
		}
		return result
	}
	return nil
}
