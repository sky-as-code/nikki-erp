package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices registers the implementation apptrait binds to core/audit.AuditSink.
func InitApplicationServices() error {
	return deps.Register(NewAuditSinkAppService)
}
