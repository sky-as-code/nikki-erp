package services

import (
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewRelationshipDomainService is handed the composable default by the onion. Plain CRUD: the type
// exists so a rule has a home the day one is needed.
func NewRelationshipDomainService(base composable.CrudDomainService) itRelationship.RelationshipDomainService {
	return &RelationshipDomainServiceImpl{CrudDomainService: base}
}

type RelationshipDomainServiceImpl struct {
	composable.CrudDomainService
}
