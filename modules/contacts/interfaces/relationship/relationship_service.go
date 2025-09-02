package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type RelationshipService interface {
	CreateRelationship(ctx crud.Context, cmd CreateRelationshipCommand) (*CreateRelationshipResult, error)
	DeleteRelationship(ctx crud.Context, cmd DeleteRelationshipCommand) (*DeleteRelationshipResult, error)
	GetRelationshipById(ctx crud.Context, query GetRelationshipByIdQuery) (*GetRelationshipByIdResult, error)
	GetRelationshipsByParty(ctx crud.Context, query GetRelationshipsByPartyQuery) (*GetRelationshipsByPartyResult, error)
	SearchRelationships(ctx crud.Context, query SearchRelationshipsQuery) (*SearchRelationshipsResult, error)
	UpdateRelationship(ctx crud.Context, cmd UpdateRelationshipCommand) (*UpdateRelationshipResult, error)
}
