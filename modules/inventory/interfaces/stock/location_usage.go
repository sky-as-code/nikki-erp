// Package stock is the port onto what Stock knows about a location.
//
// Warehouse Management decides whether a location may be suspended or archived, but that depends
// on facts only Stock holds. Asking through this contract keeps the dependency one-way and leaves
// Stock free to change how it stores any of it.
//
// The location usage port is read-only by design: nothing Warehouse does may change a quantity.
// StockTransferMovementService in transfer_movement.go publishes the movements themselves, kept
// separate so a consumer binding this contract gains no power to move anything.
package stock

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type GetLocationUsageQuery struct {
	LocationId string
}

// LocationUsage is what Stock reports about one location. The four numbers are not
// interchangeable: on-hand is what is physically there, reserved is promised to an outgoing move
// and is part of on-hand rather than additional to it, and the counts are work in flight that
// would dangle if the location went away.
type LocationUsage struct {
	OnHandQuantity    decimal.Decimal
	ReservedQuantity  decimal.Decimal
	OpenMoveCount     int
	OpenTransferCount int
}

// IsEmpty reports whether the location can be retired without stranding anything. Archiving
// requires all four clear; suspending does not — locking a rack that still holds goods is the
// point of suspension, so a caller deciding whether to suspend must not use this.
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
// Warehouse Management owns. Historical movements are deliberately not reported: a location
// referenced only by completed moves can still be archived, so counting them would block a safe
// operation forever.
type LocationUsageReadService interface {
	GetLocationUsage(ctx corectx.Context, query GetLocationUsageQuery) (*GetLocationUsageResult, error)
}
