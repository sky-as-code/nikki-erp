package models

import (
	_ "embed"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrgAuthScope = "tenant"

	OrgActionCreate      = "create"
	OrgActionDelete      = "delete"
	OrgActionUpdate      = "update"
	OrgActionView        = "view"
	OrgActionManageUsers = "manage_users"
	OrgActionSetArchived = "set_archived"
)

const (
	OrganizationSchemaName = "iam_org"

	OrgFieldId          = basemodel.FieldId
	OrgFieldAddress     = "address"
	OrgFieldDisplayName = "display_name"
	OrgFieldLegalName   = "legal_name"
	OrgFieldPhoneNumber = "phone_number"
	OrgFieldSlug        = "slug"

	OrgEdgeOrgUnits     = "org_units"
	OrgEdgeUsers        = "users"
	OrgEdgeEntitlements = "entitlements"
)

const (
	OrgUsrRelSchemaName = "iam_org_user_rel"

	OrgUsrRelFieldId     = basemodel.FieldId
	OrgUsrRelFieldUserId = "user_id"
	OrgUsrRelFieldOrgId  = "org_id"
)

// The junction carries its own id column so that a delete cascades to the user permission cache.
//
//go:embed org_user_rel.json
var orgUserRelSchemaJson string

func OrgUserRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgUserRelSchemaJson)
}

//go:embed org.json
var organizationSchemaJson string

func OrganizationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(organizationSchemaJson)
}

type Organization struct {
	basemodel.DynamicModelBase
}

func NewOrganization() *Organization {
	return &Organization{basemodel.NewDynamicModel()}
}

func NewOrganizationFrom(src dmodel.DynamicFields) *Organization {
	return &Organization{basemodel.NewDynamicModel(src)}
}

func (this Organization) MustGetId() model.Id {
	v := this.GetFieldData().GetModelId(basemodel.FieldId)
	if v == nil {
		panic("id is nil")
	}
	return *v
}

func (this Organization) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Organization) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this Organization) GetSlug() *model.Slug {
	s := this.GetFieldData().GetString(OrgFieldSlug)
	if s == nil {
		return nil
	}
	v := model.Slug(*s)
	return &v
}

func (this *Organization) SetSlug(v *model.Slug) {
	if v == nil {
		this.GetFieldData().SetString(OrgFieldSlug, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(OrgFieldSlug, &s)
}

func (this Organization) MustIsArchived() bool {
	val := this.IsArchived()
	if val == nil {
		panic(errors.New("is_archived is nil"))
	}
	return *val
}

func (this Organization) IsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

func (this Organization) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(basemodel.FieldEtag)
}

func (this *Organization) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(basemodel.FieldEtag, v)
}

func (this Organization) GetDisplayName() *model.LangJson {
	return this.GetFieldData().GetLangJson(OrgFieldDisplayName)
}

func (this *Organization) SetDisplayName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(OrgFieldDisplayName, v)
}

func (this Organization) GetLegalName() *model.LangJson {
	return this.GetFieldData().GetLangJson(OrgFieldLegalName)
}

func (this *Organization) SetLegalName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(OrgFieldLegalName, v)
}

func (this Organization) GetAddress() *model.LangJson {
	return this.GetFieldData().GetLangJson(OrgFieldAddress)
}

func (this *Organization) SetAddress(v *model.LangJson) {
	this.GetFieldData().SetLangJson(OrgFieldAddress, v)
}
