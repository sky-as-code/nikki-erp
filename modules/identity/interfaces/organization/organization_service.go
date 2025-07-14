package organization

import (
	"context"

	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error)
	DeleteOrganization(ctx context.Context, cmd DeleteOrganizationCommand) (*DeleteOrganizationResult, error)
	GetOrganizationBySlug(ctx context.Context, query GetOrganizationBySlugQuery) (*GetOrganizationBySlugResult, error)
	ListOrgStatuses(ctx context.Context, query ListOrgStatusesQuery) (*itUser.ListIdentStatusesResult, error)
	SearchOrganizations(ctx context.Context, query SearchOrganizationsQuery) (*SearchOrganizationsResult, error)
	UpdateOrganization(ctx context.Context, cmd UpdateOrganizationCommand) (*UpdateOrganizationResult, error)
}
