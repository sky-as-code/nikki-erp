package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RelationshipRepository interface {
	Create(ctx crud.Context, relationship *domain.Relationship) (*domain.Relationship, error)
}

type FindByIdParam = GetRelationshipByIdQuery
type FindByPartyParam = GetRelationshipsByPartyQuery

type FindByPartyIdsParam struct {
	PartyId       string
	TargetPartyID string
	Type          string
}
