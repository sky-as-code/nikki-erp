package external

import (
	itAuditlog "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/auditlog"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
)

// NewAuditSinkImpl points core's audit port at the Audit Log module, which the dynamic resource
// engine and every custom mutation record through instead of importing Audit Log.
//
// Audit Log lives in this binary, so the port is an in-process call joining the caller's
// transaction. A build without it leaves the port unresolvable, and nothing is recorded.
func NewAuditSinkImpl(svc itAuditlog.AuditSinkAppService) audit.AuditSink {
	return svc
}
