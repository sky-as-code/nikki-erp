package role

import "context"

type RoleService interface {
	CreateRole(ctx context.Context, cmd CreateRoleCommand) (*CreateRoleResult, error)
	// UpdateResource(ctx context.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
	// GetResourceByName(ctx context.Context, cmd GetResourceByNameCommand) (*GetResourceByNameResult, error)
	// SearchResources(ctx context.Context, query SearchResourcesCommand) (*SearchResourcesResult, error)
}
