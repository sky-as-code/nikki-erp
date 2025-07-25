package resource

import (
	"context"
)

type ResourceService interface {
	CreateResource(ctx context.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	UpdateResource(ctx context.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
	DeleteHardResource(ctx context.Context, cmd DeleteHardResourceCommand) (*DeleteHardResourceResult, error)
	GetResourceByName(ctx context.Context, query GetResourceByNameQuery) (*GetResourceByNameResult, error)
	SearchResources(ctx context.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
}
