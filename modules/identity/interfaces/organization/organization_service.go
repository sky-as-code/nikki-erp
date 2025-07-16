package organization

import (
	"context"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error)
	DeleteOrganization(ctx context.Context, cmd DeleteOrganizationCommand) (*DeleteOrganizationResult, error)
	GetOrganizationBySlug(ctx context.Context, query GetOrganizationBySlugQuery) (*GetOrganizationBySlugResult, error)
	SearchOrganizations(ctx context.Context, query SearchOrganizationsQuery) (*SearchOrganizationsResult, error)
	UpdateOrganization(ctx context.Context, cmd UpdateOrganizationCommand) (*UpdateOrganizationResult, error)
}
