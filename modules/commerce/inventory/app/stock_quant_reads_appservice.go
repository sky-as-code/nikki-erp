package app

import (
	"time"

	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The reads that let the Product UI show stock it does not own. They hang off the quant because
// what they read is quants, and they are all reads: none creates a movement, reserves anything or
// changes a balance. The batch form exists so a listing summarises a whole page in one request.

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

	// CanArchive is the reader's own verdict on its four numbers, so a caller does not restate
	// the rule and risk restating it differently.
	CanArchive bool `json:"canArchive"`
}

func (this *StockQuantApplicationServiceImpl) VariantStockSummary(
	ctx corectx.Context, query itStock.VariantStockSummaryQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	variantId := readStringField(query, paramSummaryVariantId)
	result, err := this.quantSvc.GetVariantSummaries(ctx, itStock.GetVariantSummariesQuery{VariantIds: []string{variantId}})
	if err != nil {
		return nil, err
	}
	return anyResult(toVariantSummaryJson(variantId, result.Data.Summaries[variantId])), nil
}

func (this *StockQuantApplicationServiceImpl) VariantStockSummaries(
	ctx corectx.Context, query itStock.VariantStockSummariesQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.quantSvc.GetVariantSummaries(ctx, itStock.GetVariantSummariesQuery{
		VariantIds: readStringSliceField(query, paramSummaryVariantIds),
	})
	if err != nil {
		return nil, err
	}
	// Keyed by variant id so a caller matches each summary to its row without relying on order.
	response := make(map[string]variantSummaryResponse, len(result.Data.Summaries))
	for variantId, summary := range result.Data.Summaries {
		response[variantId] = toVariantSummaryJson("", summary)
	}
	return anyResult(response), nil
}

func (this *StockQuantApplicationServiceImpl) TemplateStockSummary(
	ctx corectx.Context, query itStock.TemplateStockSummaryQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.quantSvc.GetTemplateSummary(ctx, itStock.GetTemplateSummaryQuery{
		TemplateId: readStringField(query, paramSummaryTemplateId),
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
	return anyResult(templateSummaryResponse{
		Summary:  toVariantSummaryJson("", result.Data.Summary),
		Variants: rows,
	}), nil
}

func (this *StockQuantApplicationServiceImpl) StockByWarehouse(
	ctx corectx.Context, query itStock.StockByWarehouseQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.quantSvc.GetStockByWarehouse(ctx, itStock.GetStockByWarehouseQuery{
		VariantId: readStringField(query, paramSummaryVariantId),
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
	return anyResult(rows), nil
}

func (this *StockQuantApplicationServiceImpl) StockByLocation(
	ctx corectx.Context, query itStock.StockByLocationQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.quantSvc.GetStockByLocation(ctx, itStock.GetStockByLocationQuery{
		VariantId:   readStringField(query, paramSummaryVariantId),
		WarehouseId: readStringField(query, paramSummaryWarehouse),
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
	return anyResult(rows), nil
}

// ProductUsage reports what would be stranded if a variant were archived now, so a UI can explain
// a refusal before the user attempts it. Not the enforcement point: the archive is guarded
// independently in the product service.
func (this *StockQuantApplicationServiceImpl) ProductUsage(
	ctx corectx.Context, query itStock.ProductUsageQuery,
) (*itStock.ProductStockReadResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionReadProductStock, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	result, err := this.quantSvc.GetProductUsage(ctx, itStock.GetProductUsageQuery{
		VariantId: readStringField(query, paramSummaryVariantId),
	})
	if err != nil {
		return nil, err
	}
	usage := result.Data.Usage
	return anyResult(productUsageResponse{
		OnHandQuantity:    usage.OnHandQuantity.String(),
		ReservedQuantity:  usage.ReservedQuantity.String(),
		OpenMoveCount:     usage.OpenMoveCount,
		OpenTransferCount: usage.OpenTransferCount,
		CanArchive:        usage.IsEmpty(),
	}), nil
}

func toVariantSummaryJson(variantId string, summary itStock.VariantStockSummary) variantSummaryResponse {
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
