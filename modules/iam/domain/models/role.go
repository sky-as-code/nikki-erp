package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RoleSchemaName = "iam_role"

	RoleFieldId                = "id"
	RoleFieldName              = "name"
	RoleFieldDescription       = "description"
	RoleFieldOwnerUserId       = "owner_user_id"
	RoleFieldOwnerGroupId      = "owner_group_id"
	RoleFieldIsPrivate         = "is_private"
	RoleFieldIsRequestable     = "is_requestable"
	RoleFieldIsRequiredAttach  = "is_required_attachment"
	RoleFieldIsRequiredComment = "is_required_comment"
	RoleFieldOrgId             = "org_id"

	RoleEdgeRoleRequests   = "role_requests"
	RoleEdgeEntitlements   = "entitlements"
	RoleEdgeAssignedGroups = "assigned_groups"
	RoleEdgeAssignedUsers  = "assigned_users"
	RoleEdgeOwnerGroup     = "owner_group"
	RoleEdgeOwnerUser      = "owner_user"
)

//go:embed role.json
var roleSchemaJson string

func RoleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(roleSchemaJson)
}

type Role struct {
	basemodel.DynamicModelBase
}

func NewRole() *Role {
	return &Role{basemodel.NewDynamicModel()}
}

func NewRoleFrom(src dmodel.DynamicFields) *Role {
	return &Role{basemodel.NewDynamicModel(src)}
}

func (this Role) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Role) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this Role) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(RoleFieldOrgId)
}

func (this *Role) SetOrgId(id *model.Id) {
	this.GetFieldData().SetModelId(RoleFieldOrgId, id)
}

func (this Role) IsPrivate() *bool {
	return this.GetFieldData().GetBool(RoleFieldIsPrivate)
}

func (this *Role) SetIsPrivate(v *bool) {
	this.GetFieldData().SetBool(RoleFieldIsPrivate, v)
}
