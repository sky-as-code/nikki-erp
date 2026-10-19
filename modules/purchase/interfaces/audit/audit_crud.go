// Package audit declares the CRUD layers of the purchase audit trail.
package audit

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// AuditEventCrudRepository reads and writes purchase audit events.
type AuditEventCrudRepository interface {
	composable.CrudRepository
}

// AuditEventCrudDomainService is the CRUD of the resource. Every client write is refused: an event is written
// only by the audit helper, inside the transaction of the transition it records.
type AuditEventCrudDomainService interface {
	composable.CrudDomainService
}

// AuditEventCrudApplicationService is the authorized surface the REST handler serves.
type AuditEventCrudApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateAuditEventCommand      = composable.CreateCommand
	UpdateAuditEventCommand      = composable.UpdateCommand
	DeleteAuditEventCommand      = composable.DeleteCommand
	SetAuditEventArchivedCommand = composable.SetArchivedCommand
	GetAuditEventByIdQuery       = composable.GetByIdQuery
	SearchAuditEventsQuery       = composable.SearchQuery
	AuditEventExistsQuery        = composable.ExistsQuery
)
