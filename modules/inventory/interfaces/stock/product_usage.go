package stock

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The read-only port onto whether a product may be retired.
//
// Product Management decides whether a variant or template may be archived, but that decision
// depends on facts only Stock holds: what is on the shelf, what is promised, and what is in
// flight. Rather than have the product services read stock tables directly, they ask through this
// contract — the same shape as LocationUsageReadService in location_usage.go, for the same reason.
//
// Two ports rather than one generic one: the queries differ (a variant id, not a location id) and
// merging them would mean a parameter object with a mode flag, which reads worse at both call
// sites. See CR §14 and §19.2.

// ProductUsage is what Stock reports about one variant.
//
// The four values are exactly the ones the requirement names as archive blockers (CR §14, §14.1).
// Reserved is part of OnHand rather than additional to it, so the two are not summed.
type ProductUsage struct {
	OnHandQuantity    decimal.Decimal
	ReservedQuantity  decimal.Decimal
	OpenMoveCount     int
	OpenTransferCount int
}

// IsEmpty reports whether the variant can be archived without stranding anything.
//
// Historical movement is deliberately not consulted. A variant referenced only by completed moves
// archives fine — the records keep resolving it — so counting history would block a safe operation
// forever (CR §14.2, AC-PROD-INT-031, TS-PROD-11).
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
	// variants before archiving any, so the guard reads them together rather than one per
	// iteration of a loop that is already writing (CR §14.3, TS-PROD-12).
	VariantIds []string
}

type GetProductUsageBatchResultData struct {
	// Usages is keyed by variant id. A variant with no stock is present with a zero usage
	// rather than absent.
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
