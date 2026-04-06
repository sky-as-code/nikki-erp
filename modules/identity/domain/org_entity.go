package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrganizationSchemaName = "identity.organization"

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
	OrgUsrRelSchemaName = "identity.org_user_rel"

	OrgUsrRelFieldId     = basemodel.FieldId
	OrgUsrRelFieldUserId = "user_id"
	OrgUsrRelFieldOrgId  = "org_id"
)

func OrgUserRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(OrgUsrRelSchemaName).
		TableName("ident_org_user_rel").
		ShouldBuildDb().
		// Add `id` column to cascade delete to user_permissions table
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(OrgUsrRelFieldOrgId, OrgUsrRelFieldUserId).
		Field(
			dmodel.DefineField().
				Name(OrgUsrRelFieldOrgId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(OrgUsrRelFieldUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		)
}

func OrganizationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(OrganizationSchemaName).
		Label(model.LangJson{"en-US": "Organization"}).
		TableName("ident_organizations").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(OrgFieldAddress).
				Label(model.LangJson{"en-US": "Address"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldDisplayName).
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldLegalName).
				Label(model.LangJson{"en-US": "Legal Name"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldPhoneNumber).
				Label(model.LangJson{"en-US": "Phone"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldSlug).
				Label(model.LangJson{"en-US": "Slug"}).
				DataType(dmodel.FieldDataTypeSlug()).
				RequiredForCreate().
				Unique(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(OrgEdgeUsers).
				ManyToMany(UserSchemaName, OrgUsrRelSchemaName, "org").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(OrgEdgeOrgUnits).
				Label(model.LangJson{"en-US": "Organizational Units"}).
				Existing(OrganizationalUnitSchemaName, OrgUnitEdgeOrg),
		).
		EdgeFrom(
			dmodel.Edge(OrgEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeOrg),
		)
}

type Organization struct {
	fields dmodel.DynamicFields
}

func NewOrganization() *Organization {
	return &Organization{fields: make(dmodel.DynamicFields)}
}

func NewOrganizationFrom(src dmodel.DynamicFields) *Organization {
	return &Organization{fields: src}
}

func (this Organization) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Organization) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Organization) GetId() *model.Id {
	return this.fields.GetModelId(basemodel.FieldId)
}

func (this *Organization) SetId(v *model.Id) {
	this.fields.SetModelId(basemodel.FieldId, v)
}

func (this Organization) GetSlug() *model.Slug {
	s := this.fields.GetString(OrgFieldSlug)
	if s == nil {
		return nil
	}
	v := model.Slug(*s)
	return &v
}

func (this *Organization) SetSlug(v *model.Slug) {
	if v == nil {
		this.fields.SetString(OrgFieldSlug, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(OrgFieldSlug, &s)
}

func (this Organization) IsArchived() bool {
	isArchived := this.fields.GetBool(basemodel.FieldIsArchived)
	if isArchived == nil {
		return false
	}
	return *isArchived
}

func (this Organization) GetEtag() *model.Etag {
	return this.fields.GetEtag(basemodel.FieldEtag)
}

func (this *Organization) SetEtag(v *model.Etag) {
	this.fields.SetEtag(basemodel.FieldEtag, v)
}

func (this Organization) GetDisplayName() *string {
	return this.fields.GetString(OrgFieldDisplayName)
}
