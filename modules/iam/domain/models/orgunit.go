package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrgUnitAuthScope = "tenant"

	OrgUnitActionCreate      = "create"
	OrgUnitActionDelete      = "delete"
	OrgUnitActionUpdate      = "update"
	OrgUnitActionView        = "view"
	OrgUnitActionMove        = "move"
	OrgUnitActionManageUsers = "manage_users"
)

const (
	OrganizationalUnitSchemaName = "iam_orgunit"

	OrgUnitFieldId          = basemodel.FieldId
	OrgUnitFieldName        = "name"
	OrgUnitFieldDescription = "description"
	OrgUnitFieldPath        = "path"
	OrgUnitFieldParentId    = "parent_id"
	OrgUnitFieldOrgId       = "org_id"
	OrgUnitFieldUsers       = "users"

	OrgUnitEdgeChildren     = "children"
	OrgUnitEdgeParent       = "parent"
	OrgUnitEdgeEntitlements = "entitlements"
	OrgUnitEdgeOrg          = "org"
)

//go:embed orgunit.json
var orgUnitSchemaJson string

func OrganizationalUnitSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgUnitSchemaJson)
}

type OrganizationalUnit struct {
	basemodel.DynamicModelBase
}

func NewOrganizationalUnit() *OrganizationalUnit {
	return &OrganizationalUnit{basemodel.NewDynamicModel()}
}

func NewOrganizationalUnitFrom(src dmodel.DynamicFields) *OrganizationalUnit {
	return &OrganizationalUnit{basemodel.NewDynamicModel(src)}
}

func (this OrganizationalUnit) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *OrganizationalUnit) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this OrganizationalUnit) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(OrgUnitFieldOrgId)
}

func (this *OrganizationalUnit) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(OrgUnitFieldOrgId, v)
}

func (this OrganizationalUnit) GetParentId() *model.Id {
	return this.GetFieldData().GetModelId(OrgUnitFieldParentId)
}

func (this *OrganizationalUnit) SetParentId(v *model.Id) {
	this.GetFieldData().SetModelId(OrgUnitFieldParentId, v)
}

func (this OrganizationalUnit) GetPath() []string {
	return this.GetFieldData().GetStrings(OrgUnitFieldPath)
}

func (this *OrganizationalUnit) SetPath(path []string) {
	this.GetFieldData().SetStrings(OrgUnitFieldPath, path)
}
