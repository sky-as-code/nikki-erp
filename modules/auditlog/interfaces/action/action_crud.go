// Package action declares the layers of the auditlog_action resource: one row per audited action.
//
// The rows are written only through the core/audit.AuditSink port. Over REST the resource is
// read-only: the built-in search and get for administrators, and the per-record history any
// reader of the target record may see.
package action

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type ActionRepository interface {
	composable.CrudRepository
}

type ActionDomainService interface {
	composable.CrudDomainService
}

type ActionApplicationService interface {
	composable.CrudApplicationService

	// GetHistory lists the entries of one record, newest first, to a caller who may read that
	// record. Values of fields the caller may not read are nulled and listed in masked_fields.
	GetHistory(ctx corectx.Context, query GetHistoryQuery) (*GetHistoryResult, error)
}

// GetHistoryQuery carries org_id, resource_code, resource_id, page and size.
type GetHistoryQuery = dmodel.DynamicFields

type GetHistoryResult = dyn.OpResult[HistoryResultData]

type HistoryResultData struct {
	Items []HistoryItem `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type HistoryItem struct {
	Id           string          `json:"id"`
	Action       string          `json:"action"`
	CreatedAt    string          `json:"created_at"`
	RequestId    *string         `json:"request_id"`
	Actor        HistoryActor    `json:"actor"`
	Changes      []HistoryChange `json:"changes"`
	MaskedFields []string        `json:"masked_fields"`
}

type HistoryActor struct {
	Id           *string `json:"id"`
	Type         string  `json:"type"`
	DisplayName  *string `json:"display_name"`
	AvatarFileId *string `json:"avatar_file_id"`
}

type HistoryChange struct {
	Field string `json:"field"`
	Old   any    `json:"old"`
	New   any    `json:"new"`
	// Labels names the records an edge field's ids point at, by id, as the caller may read them.
	// Absent for a non-edge field; an id it cannot name is left out and shown raw.
	Labels map[string]string `json:"labels,omitempty"`
}
