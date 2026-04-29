package resource

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type ResourceDomainService interface {
	CreateResource(ctx corectx.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	DeleteResource(ctx corectx.Context, cmd DeleteResourceCommand) (*DeleteResourceResult, error)
	ResourceExists(ctx corectx.Context, query ResourceExistsQuery) (*ResourceExistsResult, error)
	GetResource(ctx corectx.Context, query GetResourceQuery) (*dyn.OpResult[domain.Resource], error)
	SearchResources(ctx corectx.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
	UpdateResource(ctx corectx.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
}

type ResourceAppService interface {
	CreateResource(ctx corectx.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	DeleteResource(ctx corectx.Context, cmd DeleteResourceCommand) (*DeleteResourceResult, error)
	ResourceExists(ctx corectx.Context, query ResourceExistsQuery) (*ResourceExistsResult, error)
	GetResource(ctx corectx.Context, query GetResourceQuery) (*GetResourceResult, error)
	SearchResources(ctx corectx.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
	UpdateResource(ctx corectx.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
}
