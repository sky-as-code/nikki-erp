package stock

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The read-only port onto what Stock knows about a product. Product displays these numbers and
// stores none of them: nothing here writes, and every quantity-changing operation stays behind a
// stock command.
//
// The summaries are computed on each read from quants and moves — there is no summary table or
// materialized projection — so they cannot go stale or become a second source of truth.

// MaxSummaryVariants bounds how many variants one batch summary call resolves, so a caller
// passing an unbounded id list cannot turn one summary into an unbounded scan.
const MaxSummaryVariants = 500

// VariantStockSummary is what Stock reports about one variant, in its inventory UoM.
//
// The quantities are not interchangeable. OnHand is what is physically present. Reserved is the
// part of OnHand already promised to an outgoing move — included in OnHand, never additional to
// it. Available is OnHand minus Reserved. Forecasted projects confirmed incoming and outgoing
// movement onto today's balance. InTransit has left one place and not yet arrived at the next.
type VariantStockSummary struct {
	OnHand     decimal.Decimal
	Reserved   decimal.Decimal
	Available  decimal.Decimal
	Forecasted decimal.Decimal
	InTransit  decimal.Decimal

	// LocationCount and WarehouseCount are how many distinct places hold a non-zero balance.
	LocationCount  int
	WarehouseCount int

	// BaseUomId is the UoM the quantities are expressed in. Nil when the variant has no stock
	// anywhere, since there is then no quant to read it from.
	BaseUomId *model.Id

	// LastMovementAt is when stock for this variant last actually moved. Nil when it never has.
	LastMovementAt *time.Time

	// Truncated reports that the read hit its bound, so the numbers above are a partial total. A
	// caller must not present a truncated summary as authoritative.
	Truncated bool
}

// WarehouseStockRow is one warehouse's holding of a variant. WarehouseId is nil for stock at
// locations belonging to no warehouse — vendor, customer, transit and inventory-loss locations
// all legitimately have none.
type WarehouseStockRow struct {
	WarehouseId   *model.Id
	WarehouseCode string
	WarehouseName string

	// WarehouseStatus lets the UI badge a suspended warehouse. Its stock is still shown:
	// suspension governs whether new operations may select it, not whether its contents exist.
	WarehouseStatus string

	OnHand    decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

// LocationStockRow is one location's holding of a variant.
type LocationStockRow struct {
	LocationId   model.Id
	LocationCode string
	LocationName string

	// LocationStatus lets the UI badge a suspended location, which keeps its stock and keeps being
	// displayed.
	LocationStatus string

	WarehouseId   *model.Id
	WarehouseCode string

	OnHand    decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

// TemplateVariantStockRow is one row of a template's variant breakdown.
type TemplateVariantStockRow struct {
	VariantId      model.Id
	Sku            string
	CombinationKey string

	Summary VariantStockSummary
}

type GetVariantSummariesQuery struct {
	// VariantIds is the batch to resolve: a listing summarises its whole page in one request
	// rather than one per row, avoiding an N+1.
	VariantIds []string
}

type GetVariantSummariesResultData struct {
	// Summaries is keyed by variant id. A variant with no stock is present with a zero summary
	// rather than absent.
	Summaries map[string]VariantStockSummary
}

type GetVariantSummariesResult = dyn.OpResult[GetVariantSummariesResultData]

type GetTemplateSummaryQuery struct {
	TemplateId string
}

// GetTemplateSummaryResultData carries the aggregate and the rows it was computed from. The
// aggregate is the sum of the variants, never a quantity of its own: a template has no quants and
// must never acquire any.
type GetTemplateSummaryResultData struct {
	Summary  VariantStockSummary
	Variants []TemplateVariantStockRow
}

type GetTemplateSummaryResult = dyn.OpResult[GetTemplateSummaryResultData]

type GetStockByWarehouseQuery struct {
	VariantId string
}

type GetStockByWarehouseResultData struct {
	Rows []WarehouseStockRow
}

type GetStockByWarehouseResult = dyn.OpResult[GetStockByWarehouseResultData]

type GetStockByLocationQuery struct {
	VariantId string

	// WarehouseId narrows the rows to one warehouse. Empty returns every location.
	WarehouseId string
}

type GetStockByLocationResultData struct {
	Rows []LocationStockRow
}

type GetStockByLocationResult = dyn.OpResult[GetStockByLocationResultData]

// StockProductSummaryReader is what Product reads to show stock without owning it. Methods take
// variant ids because a variant is the only thing a quant references; GetTemplateSummary is the
// exception and resolves the template to its variants before reading anything.
type StockProductSummaryReader interface {
	// GetVariantSummaries resolves a batch of variants in one call.
	GetVariantSummaries(
		ctx corectx.Context, query GetVariantSummariesQuery,
	) (*GetVariantSummariesResult, error)

	// GetTemplateSummary aggregates a template's variants and reports the breakdown alongside.
	GetTemplateSummary(
		ctx corectx.Context, query GetTemplateSummaryQuery,
	) (*GetTemplateSummaryResult, error)

	// GetStockByWarehouse groups one variant's stock by the warehouse holding it.
	GetStockByWarehouse(
		ctx corectx.Context, query GetStockByWarehouseQuery,
	) (*GetStockByWarehouseResult, error)

	// GetStockByLocation lists one variant's stock per location.
	GetStockByLocation(
		ctx corectx.Context, query GetStockByLocationQuery,
	) (*GetStockByLocationResult, error)
}
