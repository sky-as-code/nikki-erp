package entitlement

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_entitlement resource, served at /v1/iam/iam_entitlement.

type EntitlementCrudRepository interface {
	dynamicresource.CrudRepository
}

// EntitlementCrudDomainService is the CRUD of the resource plus its rules: a create validates the
// scope and derives the expression, and every change to what an entitlement grants rebuilds the
// permission cache of its role's holders and is audited.
type EntitlementCrudDomainService interface {
	dynamicresource.CrudDomainService

	// The DeferRebuild variants skip the role rebuild, so a caller changing several entitlements
	// of one role pays for a single rebuild at the end. The caller owns that rebuild.
	CreateDeferRebuild(ctx corectx.Context, cmd CreateEntitlementCommand) (*CreateEntitlementResult, error)
	UpdateDeferRebuild(ctx corectx.Context, cmd UpdateEntitlementCommand) (*UpdateEntitlementResult, error)
	DeleteDeferRebuild(ctx corectx.Context, cmd DeleteEntitlementCommand) (*DeleteEntitlementResult, error)
}

type EntitlementCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateEntitlementCommand = dynamicresource.CreateCommand
	CreateEntitlementResult  = dynamicresource.CreateResult
	UpdateEntitlementCommand = dynamicresource.UpdateCommand
	UpdateEntitlementResult  = dynamicresource.MutateResult
	DeleteEntitlementCommand = dynamicresource.DeleteCommand
	DeleteEntitlementResult  = dynamicresource.MutateResult
)
