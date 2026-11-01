package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_orgunit resource, served at /v1/iam/iam_orgunit.

type OrgUnitCrudRepository interface {
	dynamicresource.CrudRepository
}

// OrgUnitCrudDomainService is the CRUD of the resource plus its own rules: a create derives the
// unit's path from its parent, and the unit's members are the users whose org_unit_id names it.
type OrgUnitCrudDomainService interface {
	dynamicresource.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
}

type OrgUnitCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
}

type (
	CreateOrgUnitCommand = dynamicresource.CreateCommand
	CreateOrgUnitResult  = dynamicresource.CreateResult

	ManageOrgUnitUsersCommand = dynamicresource.ManageM2mCommand
	ManageOrgUnitUsersResult  = dynamicresource.MutateResult
)
