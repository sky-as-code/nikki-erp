package relationship

import (
	"context"
)

type RelationshipService interface {
	CreateRelationship(ctx context.Context, cmd CreateRelationshipCommand) (*CreateRelationshipResult, error)
	DeleteRelationship(ctx context.Context, cmd DeleteRelationshipCommand) (*DeleteRelationshipResult, error)
	GetRelationshipById(ctx context.Context, query GetRelationshipByIdQuery) (*GetRelationshipByIdResult, error)
	GetRelationshipsByParty(ctx context.Context, query GetRelationshipsByPartyQuery) (*GetRelationshipsByPartyResult, error)
	SearchRelationships(ctx context.Context, query SearchRelationshipsQuery) (*SearchRelationshipsResult, error)
	UpdateRelationship(ctx context.Context, cmd UpdateRelationshipCommand) (*UpdateRelationshipResult, error)
}
