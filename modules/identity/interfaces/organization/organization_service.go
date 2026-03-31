package organization

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type OrganizationService interface {
	CreateOrg(ctx corectx.Context, cmd CreateOrgCommand) (*CreateOrgResult, error)
	DeleteOrg(ctx corectx.Context, cmd DeleteOrgCommand) (*DeleteOrgResult, error)
	GetOrg(ctx corectx.Context, query GetOrgQuery) (*GetOrgResult, error)
	OrgExists(ctx corectx.Context, query OrgExistsQuery) (*OrgExistsResult, error)
	ManageOrgUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)
	SearchOrgs(ctx corectx.Context, query SearchOrgsQuery) (*SearchOrgsResult, error)
	SetOrgIsArchived(ctx corectx.Context, cmd SetOrgIsArchivedCommand) (*SetOrgIsArchivedResult, error)
	UpdateOrg(ctx corectx.Context, cmd UpdateOrgCommand) (*UpdateOrgResult, error)
}
