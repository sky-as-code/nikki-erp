package stock

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The read-only port onto what Stock knows about a product, for the Product UI to display.
//
// Product is an entry point onto stock, not an owner of it: it shows these numbers and links to
// the stock capability that produces them, and it stores none of them. Nothing here writes, and
// nothing here gives Product a way to change a quantity — every quantity-changing operation stays
// behind a stock command. See CR §4.1, §6.2 and PROD-INT-INV-004..006.
//
// The summaries are computed rather than stored. There is no product_stock_summary table and no
// materialized projection: the numbers are derived from quants and moves on each read, so they
// cannot go stale and cannot become a second source of truth (CR §21.4, PROD-INT-INV-022).

// MaxSummaryVariants bounds how many variants one batch summary call resolves.
//
// A product listing is a page of rows, so a few hundred is already past any real page size. The
// bound exists so that a caller passing an unbounded id list cannot turn one summary into an
// unbounded scan.
const MaxSummaryVariants = 500

// VariantStockSummary is what Stock reports about one variant, in its inventory UoM.
//
// The quantities answer different questions and are not interchangeable. OnHand is what is
// physically present. Reserved is the part of OnHand already promised to an outgoing move — it is
// included in OnHand, never additional to it. Available is OnHand minus Reserved (BR §4.2.2.3).
// Forecasted projects confirmed incoming and outgoing movement onto today's balance
// (BR §4.2.13.3). InTransit is stock that has left one place and not yet arrived at the next.
type VariantStockSummary struct {
	OnHand     decimal.Decimal
	Reserved   decimal.Decimal
	Available  decimal.Decimal
	Forecasted decimal.Decimal
	InTransit  decimal.Decimal

	// LocationCount and WarehouseCount are how many distinct places hold a non-zero balance,
	// which is what the Product UI shows as "Number of Locations" (CR §6.1).
	LocationCount  int
	WarehouseCount int

	// BaseUomId is the UoM the quantities are expressed in. Nil when the variant has no stock
	// anywhere, since there is then no quant to read it from.
	BaseUomId *model.Id

	// LastMovementAt is when stock for this variant last actually moved. Nil when it never has.
	LastMovementAt *time.Time

	// Truncated reports that the read hit its bound and the numbers above are therefore a
	// partial total. A caller must not present a truncated summary as authoritative: a wrong
	// on-hand shown as if it were right is worse than a visibly incomplete one.
	Truncated bool
}

// WarehouseStockRow is one warehouse's holding of a variant (CR §9.1).
//
// WarehouseId is nil for stock held at locations that belong to no warehouse — vendor, customer,
// transit and inventory-loss locations all legitimately have none (overlap CR §8, §8.2).
type WarehouseStockRow struct {
	WarehouseId   *model.Id
	WarehouseCode string
	WarehouseName string

	// WarehouseStatus is carried so the UI can badge a suspended warehouse. Stock at one is
	// still shown: suspension governs whether new operations may select it, not whether its
	// contents exist (CR §9.5, AC-PROD-INT-018).
	WarehouseStatus string

	OnHand    decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

// LocationStockRow is one location's holding of a variant (CR §9.2).
type LocationStockRow struct {
	LocationId   model.Id
	LocationCode string
	LocationName string

	// LocationStatus lets the UI badge a suspended location, which keeps its stock and keeps
	// being displayed (CR §9.4, AC-PROD-INT-017, TS-PROD-05).
	LocationStatus string

	WarehouseId   *model.Id
	WarehouseCode string

	OnHand    decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

// TemplateVariantStockRow is one row of a template's variant breakdown (CR §5.3, §24).
type TemplateVariantStockRow struct {
	VariantId      model.Id
	Sku            string
	CombinationKey string

	Summary VariantStockSummary
}

type GetVariantSummariesQuery struct {
	// VariantIds is the batch to resolve. Batching is the point of this call: a product listing
	// summarises its whole page in one request rather than one per row, which is the N+1 the
	// requirement forbids (CR §8.4, AC-PROD-INT-035).
	VariantIds []string
}

type GetVariantSummariesResultData struct {
	// Summaries is keyed by variant id. A variant with no stock is present with a zero summary
	// rather than absent, so a caller need not distinguish "no stock" from "not asked for".
	Summaries map[string]VariantStockSummary
}

type GetVariantSummariesResult = dyn.OpResult[GetVariantSummariesResultData]

type GetTemplateSummaryQuery struct {
	TemplateId string
}

// GetTemplateSummaryResultData carries the aggregate and the rows it was computed from.
//
// The aggregate is the sum of the variants, never a quantity of its own: a template has no quants
// and must never acquire any (CR §5.2, PROD-INT-INV-002, PROD-INT-INV-003, TS-PROD-02).
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

	// WarehouseId narrows the rows to one warehouse, for the drill-down from the by-warehouse
	// view into the locations inside it (CR §9.2). Empty returns every location.
	WarehouseId string
}

type GetStockByLocationResultData struct {
	Rows []LocationStockRow
}

type GetStockByLocationResult = dyn.OpResult[GetStockByLocationResultData]

// StockProductSummaryReader is what Product reads to show stock without owning it.
//
// Every method takes variant ids, never a template id: a variant is the concrete stocked product
// and the only thing a quant references (CR §5, PROD-INT-INV-001). GetTemplateSummary is the one
// exception in its argument, and it resolves the template to its variants before reading anything.
type StockProductSummaryReader interface {
	// GetVariantSummaries resolves a batch of variants in one call.
	GetVariantSummaries(
		ctx corectx.Context, query GetVariantSummariesQuery,
	) (*GetVariantSummariesResult, error)

	// GetTemplateSummary aggregates a template's variants, and reports the breakdown alongside.
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
