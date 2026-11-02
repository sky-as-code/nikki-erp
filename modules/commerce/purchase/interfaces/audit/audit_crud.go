// Package audit declares the CRUD layers of the purchase audit trail.
package audit

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// AuditEventCrudRepository reads and writes purchase audit events.
type AuditEventCrudRepository interface {
	dynamicresource.CrudRepository
}

// AuditEventCrudDomainService is the CRUD of the resource. Every client write is refused: an event is written
// only by the audit helper, inside the transaction of the transition it records.
type AuditEventCrudDomainService interface {
	dynamicresource.CrudDomainService
}

// AuditEventCrudApplicationService is the authorized surface the REST handler serves.
type AuditEventCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateAuditEventCommand      = dynamicresource.CreateCommand
	UpdateAuditEventCommand      = dynamicresource.UpdateCommand
	DeleteAuditEventCommand      = dynamicresource.DeleteCommand
	SetAuditEventArchivedCommand = dynamicresource.SetArchivedCommand
	GetAuditEventByIdQuery       = dynamicresource.GetByIdQuery
	SearchAuditEventsQuery       = dynamicresource.SearchQuery
	AuditEventExistsQuery        = dynamicresource.ExistsQuery
)
