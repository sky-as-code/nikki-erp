package organization

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_org resource, served at /v1/iam/iam_org.

type OrgCrudRepository interface {
	composable.CrudRepository
}

// OrgCrudDomainService is the CRUD of the resource plus its own rules: a create seeds the
// organization's settings, a delete asks every dependant module whether the org is still in use.
type OrgCrudDomainService interface {
	composable.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)

	// GetUserOrgs lists the unarchived organizations a user belongs to, ordered by display name.
	// It is the authentication path's read and asserts nothing.
	GetUserOrgs(ctx corectx.Context, userId model.Id) (*GetUserOrgsResult, error)
}

type OrgCrudApplicationService interface {
	composable.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)
}

type (
	CreateOrgCommand = composable.CreateCommand
	CreateOrgResult  = composable.CreateResult
	DeleteOrgCommand = composable.DeleteCommand
	DeleteOrgResult  = composable.MutateResult

	ManageOrgUsersCommand = composable.ManageM2mCommand
	ManageOrgUsersResult  = composable.MutateResult

	GetUserOrgsResult = composable.SearchResult
)
