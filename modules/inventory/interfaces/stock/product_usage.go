package stock

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The read-only port onto whether a product may be retired. Product Management owns the archive
// decision but it depends on facts only Stock holds, so it asks through this contract rather than
// reading stock tables. Kept separate from LocationUsageReadService because the queries differ by
// key; merging them would need a mode flag.

// ProductUsage is what Stock reports about one variant: the four archive blockers. Reserved is
// part of OnHand rather than additional to it, so the two are not summed.
type ProductUsage struct {
	OnHandQuantity    decimal.Decimal
	ReservedQuantity  decimal.Decimal
	OpenMoveCount     int
	OpenTransferCount int
}

// IsEmpty reports whether the variant can be archived without stranding anything. Historical
// movement is deliberately not consulted: a variant referenced only by completed moves archives
// fine, so counting history would block a safe operation forever.
func (this ProductUsage) IsEmpty() bool {
	return this.OnHandQuantity.IsZero() &&
		this.ReservedQuantity.IsZero() &&
		this.OpenMoveCount == 0 &&
		this.OpenTransferCount == 0
}

type GetProductUsageQuery struct {
	VariantId string
}

type GetProductUsageResultData struct {
	Usage ProductUsage
}

type GetProductUsageResult = dyn.OpResult[GetProductUsageResultData]

type GetProductUsageBatchQuery struct {
	// VariantIds is the whole set to check. Archiving a template must clear every one of its
	// variants before archiving any, so the guard reads them together rather than once per
	// iteration of a loop that is already writing.
	VariantIds []string
}

type GetProductUsageBatchResultData struct {
	// Usages is keyed by variant id. A variant with no stock is present with a zero usage rather
	// than absent.
	Usages map[string]ProductUsage
}

type GetProductUsageBatchResult = dyn.OpResult[GetProductUsageBatchResultData]

// StockProductUsageReader answers what Stock holds for a variant, for the lifecycle decisions
// Product Management owns.
type StockProductUsageReader interface {
	// GetProductUsage reports one variant's blocking stock.
	GetProductUsage(ctx corectx.Context, query GetProductUsageQuery) (*GetProductUsageResult, error)

	// GetProductUsageBatch reports a set of variants together, for the template-wide guard.
	GetProductUsageBatch(
		ctx corectx.Context, query GetProductUsageBatchQuery,
	) (*GetProductUsageBatchResult, error)
}
