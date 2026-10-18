package models

import (
	_ "embed"

	"regexp"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
)

type ResourceOwnerType string

const (
	ResourceOwnerTypeNikki  = ResourceOwnerType("nikkierp")
	ResourceOwnerTypeCustom = ResourceOwnerType("custom")
)

const (
	ResourceSchemaName = "iam_resource"

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

//go:embed resource.json
var resourceSchemaJson string

func ResourceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(resourceSchemaJson)
}

func DefineResourceFieldCode(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^\*|[a-zA-Z0-9_-]+$`),
		}))
}

func DefineResourceFieldScope(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeEnumString([]string{
			string(c.ResourceScopeTenant), string(c.ResourceScopeOrg),
			string(c.ResourceScopeOrgUnit), string(c.ResourceScopePrivate),
		}))
}

type Resource struct {
	basemodel.DynamicModelBase
}

func NewResource() *Resource {
	return &Resource{basemodel.NewDynamicModel()}
}

func NewResourceFrom(src dmodel.DynamicFields) *Resource {
	return &Resource{basemodel.NewDynamicModel(src)}
}

func (this Resource) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *Resource) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, v)
}

func (this Resource) GetCode() *string {
	return this.GetFieldData().GetString(ResourceFieldCode)
}

func (this *Resource) SetCode(v *string) {
	this.GetFieldData().SetString(ResourceFieldCode, v)
}

func (this Resource) GetMaxScope() *c.ResourceScope {
	s := this.GetFieldData().GetString(ResourceFieldMaxScope)
	if s == nil {
		return nil
	}
	v := c.ResourceScope(*s)
	return &v
}

func (this *Resource) SetMaxScope(v *c.ResourceScope) {
	if v == nil {
		this.GetFieldData().SetString(ResourceFieldMaxScope, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ResourceFieldMaxScope, &s)
}

func (this Resource) GetMinScope() *c.ResourceScope {
	s := this.GetFieldData().GetString(ResourceFieldMinScope)
	if s == nil {
		return nil
	}
	v := c.ResourceScope(*s)
	return &v
}

func (this *Resource) SetMinScope(v *c.ResourceScope) {
	if v == nil {
		this.GetFieldData().SetString(ResourceFieldMinScope, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(ResourceFieldMinScope, &s)
}

func (this Resource) GetActions() []Action {
	if this.GetFieldData()[ResourceEdgeActions] == nil {
		return nil
	}
	rawActions := this.GetFieldData()[ResourceEdgeActions].([]dmodel.DynamicFields)
	actions := array.Map(rawActions, func(rawAction dmodel.DynamicFields) Action {
		return *NewActionFrom(rawAction)
	})
	return actions
}
