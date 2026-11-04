package context

import (
	"context"
)

var ctxKeyBatchRow = contextKey{"batch_row"}

// SetBatchRow records that ctx is now writing the row-th (1-based) record of a batch, such as a bulk
// create or an import; 0 clears it. One request writing many records needs it to tell their side
// effects apart deterministically, e.g. the operation key of each record's sequence number, so
// that a retry with the same idempotency id addresses the same rows again. The owning-goroutine
// rule of the setters applies.
func SetBatchRow(ctx Context, row int) {
	ctx.WithValue(ctxKeyBatchRow, row)
}

// BatchRowOf returns the row SetBatchRow recorded; false outside a batch.
func BatchRowOf(ctx context.Context) (int, bool) {
	if ctx == nil {
		return 0, false
	}
	row, ok := ctx.Value(ctxKeyBatchRow).(int)
	return row, ok && row > 0
}
