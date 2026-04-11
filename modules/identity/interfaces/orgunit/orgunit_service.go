package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type OrgUnitService interface {
	CreateOrgUnit(ctx corectx.Context, cmd CreateOrgUnitCommand) (*CreateOrgUnitResult, error)
	DeleteOrgUnit(ctx corectx.Context, cmd DeleteOrgUnitCommand) (*DeleteOrgUnitResult, error)
	GetOrgUnit(ctx corectx.Context, query GetOrgUnitQuery) (*GetOrgUnitResult, error)
	OrgUnitExists(ctx corectx.Context, cmd OrgUnitExistsQuery) (*OrgUnitExistsResult, error)
	ManageOrgUnitUsers(ctx corectx.Context, cmd ManageOrgUnitUsersCommand) (*ManageOrgUnitUsersResult, error)
	SearchOrgUnits(ctx corectx.Context, query SearchOrgUnitsQuery) (*SearchOrgUnitsResult, error)
	UpdateOrgUnit(ctx corectx.Context, cmd UpdateOrgUnitCommand) (*UpdateOrgUnitResult, error)
}
