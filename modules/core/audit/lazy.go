package audit

import (
	"sync"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// LazyAuditSink resolves the port on first use rather than when its holder is built, for the same
// reason as composable's lazy storage consumer: the auditlog module's own onions are built by the
// engine that writes the entries, so injecting the port would form a cycle, and a holder built
// before auditlog's Init ran would keep a nil port for the life of the process.
//
// A binary that registers no sink records nothing: Enabled is false and Record is a no-op.
type LazyAuditSink struct {
	once sync.Once
	sink AuditSink
}

func NewLazyAuditSink() *LazyAuditSink {
	return &LazyAuditSink{}
}

func (this *LazyAuditSink) resolve() AuditSink {
	this.once.Do(func() {
		_ = deps.Invoke(func(sink AuditSink) {
			this.sink = sink
		})
	})
	return this.sink
}

// Enabled reports whether a sink is registered, so a writer can skip reading the before row and
// opening a transaction it would only need for the entry.
func (this *LazyAuditSink) Enabled() bool {
	return this.resolve() != nil
}

func (this *LazyAuditSink) Record(ctx corectx.Context, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	if sink := this.resolve(); sink != nil {
		return sink.Record(ctx, entries)
	}
	return nil
}
