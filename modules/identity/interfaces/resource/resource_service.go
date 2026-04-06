package resource

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type ResourceService interface {
	CreateResource(ctx corectx.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	DeleteResource(ctx corectx.Context, cmd DeleteResourceCommand) (*DeleteResourceResult, error)
	ResourceExists(ctx corectx.Context, query ResourceExistsQuery) (*ResourceExistsResult, error)
	GetResource(ctx corectx.Context, query GetResourceQuery) (*GetResourceResult, error)
	SearchResources(ctx corectx.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
	UpdateResource(ctx corectx.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
}
