package organization

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// The composable layers of the iam_org resource, served at /v1/iam/iam_org.

type OrgCrudRepository interface {
	dynamicresource.CrudRepository
}

// OrgCrudDomainService is the CRUD of the resource plus its own rules: a create seeds the
// organization's settings, a delete asks every dependant module whether the org is still in use.
type OrgCrudDomainService interface {
	dynamicresource.CrudDomainService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)

	// GetUserOrgs lists the ready, unarchived organizations a user belongs to, ordered by display
	// name: an org still provisioning or being deleted cannot be switched into.
	// It is the authentication path's read and asserts nothing.
	GetUserOrgs(ctx corectx.Context, userId model.Id) (*GetUserOrgsResult, error)

	// MarkDeleted sets status deleted and archives the org in one write; the org lifecycle then
	// cleans up every module and removes it with Delete.
	MarkDeleted(ctx corectx.Context, cmd DeleteOrgCommand) (*DeleteOrgResult, error)

	// SetLifecycleStatus records the outcome of the org lifecycle.
	SetLifecycleStatus(
		ctx corectx.Context, orgId model.Id, status domain.OrgStatus, message *string, vars map[string]any,
	) error
}

type OrgCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)
}

type (
	CreateOrgCommand = dynamicresource.CreateCommand
	CreateOrgResult  = dynamicresource.CreateResult
	DeleteOrgCommand = dynamicresource.DeleteCommand
	DeleteOrgResult  = dynamicresource.MutateResult
	UpdateOrgCommand = dynamicresource.UpdateCommand
	UpdateOrgResult  = dynamicresource.MutateResult

	SetOrgArchivedCommand = dynamicresource.SetArchivedCommand
	SetOrgArchivedResult  = dynamicresource.MutateResult

	ManageOrgUsersCommand = dynamicresource.ManageM2mCommand
	ManageOrgUsersResult  = dynamicresource.MutateResult

	GetUserOrgsResult = dynamicresource.SearchResult
)
