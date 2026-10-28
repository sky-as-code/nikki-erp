package auditlog

import "github.com/sky-as-code/nikki-erp/modules/core/audit"

// AuditSinkAppService is this module's implementation of the core/audit port, under its own type
// so the container holds it whether or not apptrait binds the port to it. apptrait registers it
// as audit.AuditSink in every binary that loads this module.
type AuditSinkAppService interface {
	audit.AuditSink
}
