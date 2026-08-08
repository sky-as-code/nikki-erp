package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	GroupSchemaName = "iam_group"

	GroupFieldId          = basemodel.FieldId
	GroupFieldName        = "name"
	GroupFieldDescription = "description"
	GroupFieldOwnerId     = "owner_id"

	GroupEdgeOwner                = "owner"
	GroupEdgeRoles                = "roles"
	GroupEdgeOwnRoles             = "own_roles"
	GroupEdgeUsers                = "users"
	GroupEdgeBenefitGrantRequests = "benefit_grant_requests"
)

const (
	GroupAuthScope = "org"

	GroupActionCreate      = "create"
	GroupActionDelete      = "delete"
	GroupActionUpdate      = "update"
	GroupActionView        = "view"
	GroupActionManageUsers = "manage_users"
)

const (
	GrpUsrRelSchemaName = "iam_group_user_rel"

	GrpUsrRelFieldId      = basemodel.FieldId
	GrpUsrRelFieldGroupId = "group_id"
	GrpUsrRelFieldUserId  = "user_id"
)

//go:embed group_user_rel.json
var groupUserRelSchemaJson string

func GroupUserRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(groupUserRelSchemaJson)
}

//go:embed group.json
var groupSchemaJson string

func GroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(groupSchemaJson)
}

type Group struct {
	basemodel.DynamicModelBase
}

func NewGroup() *Group {
	return &Group{basemodel.NewDynamicModel()}
}

func NewGroupFrom(src dmodel.DynamicFields) *Group {
	return &Group{basemodel.NewDynamicModel(src)}
}

func (this Group) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Group) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this Group) GetName() *string {
	return this.GetFieldData().GetString(GroupFieldName)
}

func (this *Group) SetName(v *string) {
	this.GetFieldData().SetString(GroupFieldName, v)
}

func (this Group) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(basemodel.FieldEtag)
}

func (this *Group) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(basemodel.FieldEtag, v)
}
