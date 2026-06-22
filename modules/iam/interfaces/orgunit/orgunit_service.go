package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

type OrgUnitDomainService interface {
	CreateOrgUnit(ctx corectx.Context, cmd CreateOrgUnitCommand, opts ...corecrud.ServiceCreateOptions[*domain.OrganizationalUnit]) (*CreateOrgUnitResult, error)
	DeleteOrgUnit(ctx corectx.Context, cmd DeleteOrgUnitCommand, opts ...corecrud.ServiceDeleteOptions) (*DeleteOrgUnitResult, error)
	GetOrgUnit(ctx corectx.Context, query GetOrgUnitQuery) (*dyn.OpResult[domain.OrganizationalUnit], error)
	OrgUnitExists(ctx corectx.Context, cmd OrgUnitExistsQuery) (*OrgUnitExistsResult, error)
	ManageOrgUnitUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
	SearchOrgUnits(ctx corectx.Context, query SearchOrgUnitsQuery, opts ...corecrud.ServiceSearchOptions) (*SearchOrgUnitsResult, error)
	UpdateOrgUnit(ctx corectx.Context, cmd UpdateOrgUnitCommand, opts ...corecrud.ServiceUpdateOptions[*domain.OrganizationalUnit]) (*UpdateOrgUnitResult, error)
}

type OrgUnitAppService interface {
	CreateOrgUnit(ctx corectx.Context, cmd CreateOrgUnitCommand) (*CreateOrgUnitResult, error)
	DeleteOrgUnit(ctx corectx.Context, cmd DeleteOrgUnitCommand) (*DeleteOrgUnitResult, error)
	GetOrgUnit(ctx corectx.Context, query GetOrgUnitQuery) (*GetOrgUnitResult, error)
	OrgUnitExists(ctx corectx.Context, cmd OrgUnitExistsQuery) (*OrgUnitExistsResult, error)
	ManageOrgUnitUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
	SearchOrgUnits(ctx corectx.Context, query SearchOrgUnitsQuery) (*SearchOrgUnitsResult, error)
	UpdateOrgUnit(ctx corectx.Context, cmd UpdateOrgUnitCommand) (*UpdateOrgUnitResult, error)
}
