package database

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// RunOnCommitted queues fn on tranx, or runs it at once when there is no transaction: a write
// made without one has already autocommitted. The immediate path recovers a panic the same way
// the transaction does, so a side effect never fails a write that already succeeded.
func RunOnCommitted(tranx DbTransaction, fn func(), logger logging.LoggerService) {
	if fn == nil {
		return
	}
	if tranx != nil {
		tranx.OnCommitted(fn)
		return
	}
	defer func() {
		if err := ft.RecoverPanic(recover(), "OnCommitted callback panicked"); err != nil {
			if logger == nil {
				logger = logging.Logger()
			}
			logger.Error("OnCommitted callback panicked", err)
		}
	}()
	fn()
}
