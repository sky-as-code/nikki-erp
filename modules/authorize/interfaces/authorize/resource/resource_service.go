package resource

import (
	"context"
)

type ResourceService interface {
	CreateResource(ctx context.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	UpdateResource(ctx context.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
	GetResourceByName(ctx context.Context, query GetResourceByNameQuery) (*GetResourceByNameResult, error)
	SearchResources(ctx context.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
}
