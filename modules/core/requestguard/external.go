package requestguard

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

/*
 * Copied from nikkierp/modules/identity/interfaces/permission/commands.go
 */
var getUserEntQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permission",
	Action:    "getUserEntitlements",
}

type ExtGetUserEntitlementsQuery struct {
	UserId    *model.Id `json:"user_id"`
	UserEmail *string   `json:"user_email"`
}

func (ExtGetUserEntitlementsQuery) CqrsRequestType() cqrs.RequestType {
	return getUserEntQueryType
}

type ExtGetUserEntitlementsResultData struct {
	IsOwner      bool                 `json:"is_owner"`
	Entitlements []string             `json:"entitlements"`
	OrgUnitId    *model.Id            `json:"org_unit_id"`
	OrgUnitOrgId *model.Id            `json:"org_unit_org_id"`
	UserId       model.Id             `json:"user_id"`
	UserOrgIds   []model.Id           `json:"user_org_ids"`
	User         dmodel.DynamicFields `json:"user"`
}

type ExtGetUserEntitlementsResult = dyn.OpResult[ExtGetUserEntitlementsResultData]
