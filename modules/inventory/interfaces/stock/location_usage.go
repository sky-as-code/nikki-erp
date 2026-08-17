// Package stock is the read-only port onto what Stock knows about a location.
//
// Warehouse Management decides whether a location may be suspended or archived, but that decision
// depends on facts only Stock holds: what is sitting there, what is promised to someone, and what
// is on its way. Rather than have the warehouse services read stock tables directly, they ask
// through this contract — so the dependency runs one way and Stock stays free to change how it
// stores any of it.
//
// The port is read-only by design. Nothing Warehouse does may change a quantity.
package stock

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type GetLocationUsageQuery struct {
	LocationId string
}

// LocationUsage is what Stock reports about one location.
//
// The four numbers answer different questions and are not interchangeable. On-hand is what is
// physically there. Reserved is what has been promised to an outgoing move but not yet moved — it
// is part of on-hand, not additional to it. The two counts are work in flight that would be left
// dangling if the location went away.
type LocationUsage struct {
	OnHandQuantity    decimal.Decimal
	ReservedQuantity  decimal.Decimal
	OpenMoveCount     int
	OpenTransferCount int
}

// IsEmpty reports whether the location can be retired without stranding anything.
//
// Archiving requires all four to be clear. Suspending deliberately does not: locking a rack that
// still holds goods is the whole point of suspension, so a caller checking whether to suspend must
// not use this.
func (this LocationUsage) IsEmpty() bool {
	return this.OnHandQuantity.IsZero() &&
		this.ReservedQuantity.IsZero() &&
		this.OpenMoveCount == 0 &&
		this.OpenTransferCount == 0
}

type GetLocationUsageResultData struct {
	Usage LocationUsage
}

type GetLocationUsageResult = dyn.OpResult[GetLocationUsageResultData]

// LocationUsageReadService answers what Stock holds at a location, for the lifecycle decisions
// Warehouse Management owns.
//
// Historical movements are deliberately not reported. A location referenced only by completed
// moves can still be archived — the records keep resolving it — so counting them would block a
// perfectly safe operation forever.
type LocationUsageReadService interface {
	GetLocationUsage(ctx corectx.Context, query GetLocationUsageQuery) (*GetLocationUsageResult, error)
}
