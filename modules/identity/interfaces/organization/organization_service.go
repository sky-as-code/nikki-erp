package organization

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type OrganizationService interface {
	AddRemoveUsers(ctx crud.Context, cmd AddRemoveUsersCommand) (*AddRemoveUsersResult, error)
	CreateOrganization(ctx crud.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error)
	DeleteOrganization(ctx crud.Context, cmd DeleteOrganizationCommand) (*DeleteOrganizationResult, error)
	GetOrganizationBySlug(ctx crud.Context, query GetOrganizationBySlugQuery) (*GetOrganizationBySlugResult, error)
	GetOrganizationById(ctx crud.Context, query GetOrganizationByIdQuery) (*GetOrganizationByIdResult, error)
	SearchOrganizations(ctx crud.Context, query SearchOrganizationsQuery) (*SearchOrganizationsResult, error)
	UpdateOrganization(ctx crud.Context, cmd UpdateOrganizationCommand) (*UpdateOrganizationResult, error)
	ExistsOrgById(ctx crud.Context, cmd ExistsOrgByIdCommand) (*ExistsOrgByIdResult, error)
}
