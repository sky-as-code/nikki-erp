package entitlement

import "github.com/sky-as-code/nikki-erp/modules/core/crud"

type EntitlementService interface {
	CreateEntitlement(ctx crud.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	EntitlementExists(ctx crud.Context, cmd EntitlementExistsQuery) (*EntitlementExistsResult, error)
	UpdateEntitlement(ctx crud.Context, cmd UpdateEntitlementCommand) (*UpdateEntitlementResult, error)
	DeleteEntitlementHard(ctx crud.Context, cmd DeleteEntitlementHardByIdCommand) (*DeleteEntitlementHardByIdResult, error)
	GetEntitlementById(ctx crud.Context, query GetEntitlementByIdQuery) (*GetEntitlementByIdResult, error)
	GetAllEntitlementByIds(ctx crud.Context, query GetAllEntitlementByIdsQuery) (*GetAllEntitlementByIdsResult, error)
	SearchEntitlements(ctx crud.Context, query SearchEntitlementsQuery) (*SearchEntitlementsResult, error)
}
