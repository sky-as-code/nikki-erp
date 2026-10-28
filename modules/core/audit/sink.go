// Package audit is the core port through which the platform records who did what to which record.
// It holds interfaces, DTOs and the diff rules only: the auditlog module implements the port and
// apptrait registers that implementation, so a writer never depends on the auditlog module.
package audit

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Built-in action codes. A custom action records its own code ("suspend", "manage_users").
const (
	ActionCreate    = "create"
	ActionUpdate    = "update"
	ActionDelete    = "delete"
	ActionArchive   = "archive"
	ActionUnarchive = "unarchive"
	ActionLink      = "link"
	ActionUnlink    = "unlink"
	ActionGet       = "get"
	ActionSearch    = "search"
)

// MetadataAuditExcluded is the field metadata key that keeps a field out of every diff, for a
// value that is sensitive without being typed `secret` (a token stored as a plain string).
const MetadataAuditExcluded = "audit_excluded"

// Entry is one audited action on one record (or one search, which has no record). Before and
// After hold only the changed fields, already filtered through Diff.
type Entry struct {
	Action       string
	ResourceCode string
	// ResourceId is nil for a search.
	ResourceId *string
	// OrgId is nil for a global record, such as an org-less role.
	OrgId  *model.Id
	Before dmodel.DynamicFields
	After  dmodel.DynamicFields
}

// IsEmpty reports an entry that records no change, which an update must not write.
func (this Entry) IsEmpty() bool {
	return len(this.Before) == 0 && len(this.After) == 0
}

// AuditSink stores audit entries. The actor, request id and source are read from ctx.
//
// Calls join the caller's transaction (ctx.GetDbTranx()), so an entry commits or rolls back with
// the change it describes, and a failure to record must fail the caller's action. The port asserts
// no permission: the caller's own application service has already authorized the action.
type AuditSink interface {
	Record(ctx corectx.Context, entries []Entry) error
}
