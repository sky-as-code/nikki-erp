package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_orgunit resource, served at /v1/iam/iam_orgunit.

type OrgUnitCrudRepository interface {
	composable.CrudRepository
}

// OrgUnitCrudDomainService is the CRUD of the resource plus its own rules: a create derives the
// unit's path from its parent, and the unit's members are the users whose org_unit_id names it.
type OrgUnitCrudDomainService interface {
	composable.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
}

type OrgUnitCrudApplicationService interface {
	composable.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
}

type (
	CreateOrgUnitCommand = composable.CreateCommand
	CreateOrgUnitResult  = composable.CreateResult

	ManageOrgUnitUsersCommand = composable.ManageM2mCommand
	ManageOrgUnitUsersResult  = composable.MutateResult
)
