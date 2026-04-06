package permission

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type UserPermissionRepository interface {
	RebuildUserPermission(ctx corectx.Context, userId model.Id) error
	RebuildAllUserPermissions(ctx corectx.Context) error
	GetOne(ctx corectx.Context, param GetUserPermissionParam) (*dyn.OpResult[dmodel.DynamicFields], error)
}

type GetUserPermissionParam struct {
	UserId       model.Id `json:"user_id"`
	ActionCode   string   `json:"action_code"`
	ResourceCode string   `json:"resource_code"`
	Scope        string   `json:"scope"`
	OrgId        model.Id `json:"org_id"`
	OrgUnitId    model.Id `json:"org_unit_id"`
}
