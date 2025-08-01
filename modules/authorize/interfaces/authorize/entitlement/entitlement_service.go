package entitlement

import "context"

type EntitlementService interface {
	CreateEntitlement(ctx context.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	EntitlementExists(ctx context.Context, cmd EntitlementExistsCommand) (*EntitlementExistsResult, error)
	UpdateEntitlement(ctx context.Context, cmd UpdateEntitlementCommand) (*UpdateEntitlementResult, error)
	DeleteEntitlementHard(ctx context.Context, cmd DeleteEntitlementHardByIdQuery) (*DeleteEntitlementHardByIdResult, error)
	GetEntitlementById(ctx context.Context, query GetEntitlementByIdQuery) (*GetEntitlementByIdResult, error)
	GetAllEntitlementByIds(ctx context.Context, query GetAllEntitlementByIdsQuery) (*GetAllEntitlementByIdsResult, error)
	SearchEntitlements(ctx context.Context, query SearchEntitlementsQuery) (*SearchEntitlementsResult, error)
}
