package organization

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

type OrganizationDomainService interface {
	CreateOrg(ctx corectx.Context, cmd CreateOrgCommand, opts ...corecrud.ServiceCreateOptions[*domain.Organization]) (*CreateOrgResult, error)
	DeleteOrg(ctx corectx.Context, cmd DeleteOrgCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteOrgResult, error)
	GetOrg(ctx corectx.Context, query GetOrgQuery) (*dyn.OpResult[domain.Organization], error)
	GetUserOrgs(ctx corectx.Context, query GetUserOrgsQuery) (*GetUserOrgsResult, error)
	OrgExists(ctx corectx.Context, query OrgExistsQuery) (*OrgExistsResult, error)
	ManageOrgUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)
	SearchOrgs(ctx corectx.Context, query SearchOrgsQuery, opts ...corecrud.ServiceSearchOptions) (*SearchOrgsResult, error)
	SetOrgIsArchived(ctx corectx.Context, cmd SetOrgIsArchivedCommand) (*SetOrgIsArchivedResult, error)
	UpdateOrg(ctx corectx.Context, cmd UpdateOrgCommand, opts ...corecrud.ServiceUpdateOptions[*domain.Organization]) (*UpdateOrgResult, error)
}

type OrganizationAppService interface {
	CreateOrg(ctx corectx.Context, cmd CreateOrgCommand) (*CreateOrgResult, error)
	DeleteOrg(ctx corectx.Context, cmd DeleteOrgCommand) (*DeleteOrgResult, error)
	GetOrg(ctx corectx.Context, query GetOrgQuery) (*GetOrgResult, error)
	OrgExists(ctx corectx.Context, query OrgExistsQuery) (*OrgExistsResult, error)
	ManageOrgUsers(ctx corectx.Context, cmd ManageOrgUsersCommand) (*ManageOrgUsersResult, error)
	SearchOrgs(ctx corectx.Context, query SearchOrgsQuery) (*SearchOrgsResult, error)
	SetOrgIsArchived(ctx corectx.Context, cmd SetOrgIsArchivedCommand) (*SetOrgIsArchivedResult, error)
	UpdateOrg(ctx corectx.Context, cmd UpdateOrgCommand) (*UpdateOrgResult, error)
}
