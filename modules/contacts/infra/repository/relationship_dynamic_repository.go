package repository

import (
	itRelationship "github.com/sky-as-code/nikki-erp/modules/contacts/interfaces/relationship"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewRelationshipRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewRelationshipRepository(base composable.CrudRepository) itRelationship.RelationshipRepository {
	return &RelationshipRepositoryImpl{CrudRepository: base}
}

type RelationshipRepositoryImpl struct {
	composable.CrudRepository
}
