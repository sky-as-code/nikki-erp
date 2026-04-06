package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type ResourceOwnerType string
type ResourceScope string

const (
	ResourceOwnerTypeNikki  = ResourceOwnerType("nikkierp")
	ResourceOwnerTypeCustom = ResourceOwnerType("custom")

	ResourceScopeDomain  = ResourceScope("domain")
	ResourceScopeOrg     = ResourceScope("org")
	ResourceScopeOrgUnit = ResourceScope("org_unit")
	ResourceScopePrivate = ResourceScope("private")
)

const (
	ResourceSchemaName = "authorize.resource"

	ResourceFieldId          = basemodel.FieldId
	ResourceFieldName        = "name"
	ResourceFieldCode        = "code"
	ResourceFieldDescription = "description"
	ResourceFieldOwnerType   = "owner_type"
	ResourceFieldMaxScope    = "max_scope"
	ResourceFieldMinScope    = "min_scope"

	ResourceEdgeActions      = "actions"
	ResourceEdgeEntitlements = "entitlements"
)

func ResourceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ResourceSchemaName).
		Label(model.LangJson{"en-US": "Resource"}).
		TableName("authz_resources").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(ResourceFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile("^[a-zA-Z0-9_-]+$"),
				})).
				RequiredForCreate().
				NoUpdate(),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldOwnerType).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ResourceOwnerTypeNikki), string(ResourceOwnerTypeCustom),
				})).
				Default(string(ResourceOwnerTypeNikki)).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "A resource can be owned by one of NikkiERP modules, or by a 3rd party system. " +
					"If owner_type=nikkierp, this resource is used for NikkiERP module authorization. " +
					"If owner_type=custom, this resource is used for 3rd party system authorization.",
				}),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldMaxScope).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ResourceScopeDomain), string(ResourceScopeOrg),
					string(ResourceScopeOrgUnit), string(ResourceScopePrivate),
				})).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "The largest scope of the resource. " +
					"No entitlement can be granted with a scope larger than this. ",
				}),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldMinScope).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ResourceScopeDomain), string(ResourceScopeOrg),
					string(ResourceScopeOrgUnit), string(ResourceScopePrivate),
				})).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "The smallest scope of the resource. " +
					"No entitlement can be granted with a scope less than this. ",
				}),
		).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(ResourceEdgeActions).
				Label(model.LangJson{"en-US": "Actions"}).
				Existing(ActionSchemaName, ActionEdgeResource),
		)
}

type Resource struct {
	fields dmodel.DynamicFields
}

func NewResource() *Resource {
	return &Resource{fields: make(dmodel.DynamicFields)}
}

func NewResourceFrom(src dmodel.DynamicFields) *Resource {
	return &Resource{fields: src}
}

func (this Resource) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Resource) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Resource) GetMaxScope() *ResourceScope {
	s := this.fields.GetString(ResourceFieldMaxScope)
	if s == nil {
		return nil
	}
	v := ResourceScope(*s)
	return &v
}

func (this *Resource) SetMaxScope(v *ResourceScope) {
	if v == nil {
		this.fields.SetString(ResourceFieldMaxScope, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(ResourceFieldMaxScope, &s)
}

func (this Resource) GetMinScope() *ResourceScope {
	s := this.fields.GetString(ResourceFieldMinScope)
	if s == nil {
		return nil
	}
	v := ResourceScope(*s)
	return &v
}

func (this *Resource) SetMinScope(v *ResourceScope) {
	if v == nil {
		this.fields.SetString(ResourceFieldMinScope, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(ResourceFieldMinScope, &s)
}
