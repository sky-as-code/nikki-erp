package resource

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type ResourceService interface {
	CreateResource(ctx crud.Context, cmd CreateResourceCommand) (*CreateResourceResult, error)
	UpdateResource(ctx crud.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
	DeleteResourceHard(ctx crud.Context, cmd DeleteResourceHardByNameQuery) (*DeleteResourceHardByNameResult, error)
	GetResourceByName(ctx crud.Context, query GetResourceByNameQuery) (*GetResourceByNameResult, error)
	SearchResources(ctx crud.Context, query SearchResourcesQuery) (*SearchResourcesResult, error)
}
