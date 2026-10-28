package constants

const AuditLogModuleName = "auditlog"

// ResourceAuditLogAction is the IAM resource code of an audit entry. The engine asserts
// permissions using the schema name as the resource code, so it must equal ActionSchemaName.
const ResourceAuditLogAction = "auditlog_action"

// REST contract under /v1/auditlog, shared with the frontend.
const RouteActionHistory = "/actions/history"

// History paging. A page past MaxHistorySize is clamped rather than refused: the panel asks for
// a screenful at a time and never needs more.
const (
	DefaultHistorySize = 20
	MaxHistorySize     = 100
)
