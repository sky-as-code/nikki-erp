package relationship

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type RelationshipService interface {
	CreateRelationship(ctx corectx.Context, cmd CreateRelationshipCommand) (*CreateRelationshipResult, error)
	DeleteRelationship(ctx corectx.Context, cmd DeleteRelationshipCommand) (*DeleteRelationshipResult, error)
	GetRelationship(ctx corectx.Context, query GetRelationshipQuery) (*GetRelationshipResult, error)
	SearchRelationships(ctx corectx.Context, query SearchRelationshipsQuery) (*SearchRelationshipsResult, error)
	RelationshipExists(ctx corectx.Context, query RelationshipExistsQuery) (*RelationshipExistsResult, error)
	UpdateRelationship(ctx corectx.Context, cmd UpdateRelationshipCommand) (*UpdateRelationshipResult, error)
}
