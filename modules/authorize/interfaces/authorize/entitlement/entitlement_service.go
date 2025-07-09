package entitlement

import "context"

type EntitlementService interface {
	CreateEntitlement(ctx context.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	// UpdateResource(ctx context.Context, cmd UpdateResourceCommand) (*UpdateResourceResult, error)
	// GetResourceByName(ctx context.Context, cmd GetResourceByNameCommand) (*GetResourceByNameResult, error)
	// SearchResources(ctx context.Context, query SearchResourcesCommand) (*SearchResourcesResult, error)
}
