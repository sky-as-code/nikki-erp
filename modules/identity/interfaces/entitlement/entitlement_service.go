package entitlement

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type EntitlementDomainService interface {
	CreateEntitlement(ctx corectx.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	DeleteEntitlement(ctx corectx.Context, cmd DeleteEntitlementCommand) (*DeleteEntitlementResult, error)
	EntitlementExists(ctx corectx.Context, query EntitlementExistsQuery) (*EntitlementExistsResult, error)
	GetEntitlement(ctx corectx.Context, query GetEntitlementQuery) (*dyn.OpResult[domain.Entitlement], error)
	ManageEntitlementRoles(ctx corectx.Context, cmd ManageEntitlementRolesCommand) (
		*ManageEntitlementRolesResult, error,
	)
	SearchEntitlements(ctx corectx.Context, query SearchEntitlementsQuery) (*SearchEntitlementsResult, error)
	SetEntitlementIsArchived(ctx corectx.Context, cmd SetEntitlementIsArchivedCommand) (
		*SetEntitlementIsArchivedResult, error,
	)
	UpdateEntitlement(ctx corectx.Context, cmd UpdateEntitlementCommand) (*UpdateEntitlementResult, error)
}

type EntitlementAppService interface {
	CreateEntitlement(ctx corectx.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	DeleteEntitlement(ctx corectx.Context, cmd DeleteEntitlementCommand) (*DeleteEntitlementResult, error)
	EntitlementExists(ctx corectx.Context, query EntitlementExistsQuery) (*EntitlementExistsResult, error)
	GetEntitlement(ctx corectx.Context, query GetEntitlementQuery) (*GetEntitlementResult, error)
	ManageEntitlementRoles(ctx corectx.Context, cmd ManageEntitlementRolesCommand) (
		*ManageEntitlementRolesResult, error,
	)
	SearchEntitlements(ctx corectx.Context, query SearchEntitlementsQuery) (*SearchEntitlementsResult, error)
	SetEntitlementIsArchived(ctx corectx.Context, cmd SetEntitlementIsArchivedCommand) (
		*SetEntitlementIsArchivedResult, error,
	)
	UpdateEntitlement(ctx corectx.Context, cmd UpdateEntitlementCommand) (*UpdateEntitlementResult, error)
}
