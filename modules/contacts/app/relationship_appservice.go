package app

import (
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewRelationshipApplicationService is handed the composable default by the onion: the authorized CRUD,
// nothing resource-specific yet.
func NewRelationshipApplicationService(
	base composable.CrudApplicationService,
) itRelationship.RelationshipApplicationService {
	return &RelationshipApplicationServiceImpl{CrudApplicationService: base}
}

type RelationshipApplicationServiceImpl struct {
	composable.CrudApplicationService
}
