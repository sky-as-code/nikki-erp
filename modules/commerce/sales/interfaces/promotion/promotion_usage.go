package promotion

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// UsageCounterRepository is the atomic usage counter of a promotion or a voucher (D5, BR-T04). Both
// methods require an ambient transaction, the confirm's or the cancel's, so a later failure rolls
// the count back with everything else.
type UsageCounterRepository interface {
	// TryConsumeUsage adds one use when the row still has one left, in a single conditional UPDATE:
	// two confirms racing for the last use cannot both win. False means none was left.
	TryConsumeUsage(ctx corectx.Context, id string) (bool, error)

	// ReleaseUsage gives one use back, never below zero. False means there was none to give.
	ReleaseUsage(ctx corectx.Context, id string) (bool, error)
}
