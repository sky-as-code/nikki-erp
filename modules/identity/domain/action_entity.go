package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ActionSchemaName = "authorize.action"

	ActionFieldId          = "id"
	ActionFieldName        = "name"
	ActionFieldCode        = "code"
	ActionFieldDescription = "description"
	ActionFieldResourceId  = "resource_id"

	ActionEdgeResource     = "resource"
	ActionEdgeEntitlements = "entitlements"
)

func ActionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ActionSchemaName).
		Label(model.LangJson{"en-US": "Action"}).
		TableName("authz_actions").
		CompositeUnique(ActionFieldResourceId, ActionFieldName).
		CompositeUnique(ActionFieldResourceId, ActionFieldCode).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(ActionFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			DefineActionFieldCode(ActionFieldCode).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(ActionFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(ActionFieldResourceId).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ActionEdgeResource).
				Label(model.LangJson{"en-US": "Resource"}).
				ManyToOne(ResourceSchemaName, dmodel.DynamicFields{
					ActionFieldResourceId: ResourceFieldId,
				}).
				OnDelete(dmodel.RelationCascadeNoAction),
		).
		EdgeFrom(
			dmodel.Edge(ActionEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeAction),
		)
}

func DefineActionFieldCode(fieldName string) *dmodel.FieldBuilder {
	return dmodel.DefineField().Name(fieldName).
		DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^\*|[a-zA-Z0-9_-]+$`),
		}))
}

func DefineActionFieldCodeArr(fieldName ...string) *dmodel.FieldBuilder {
	fName := ActionFieldCode
	if len(fieldName) > 0 {
		fName = fieldName[0]
	}
	return dmodel.DefineField().Name(fName).
		DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^\*|[a-zA-Z0-9_-]+$`),
		}).ArrayType())
}

type Action struct {
	fields dmodel.DynamicFields
}

func NewAction() *Action {
	return &Action{fields: make(dmodel.DynamicFields)}
}

func NewActionFrom(src dmodel.DynamicFields) *Action {
	return &Action{fields: src}
}

func (this Action) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Action) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Action) GetId() *model.Id {
	return this.fields.GetModelId(ActionFieldId)
}

func (this *Action) SetId(v *model.Id) {
	this.fields.SetModelId(ActionFieldId, v)
}

func (this Action) GetCode() *string {
	return this.fields.GetString(ActionFieldCode)
}

func (this *Action) SetCode(v *string) {
	this.fields.SetString(ActionFieldCode, v)
}

func (this Action) GetResourceId() *model.Id {
	return this.fields.GetModelId(ActionFieldResourceId)
}

func (this *Action) SetResourceId(v *model.Id) {
	this.fields.SetModelId(ActionFieldResourceId, v)
}

func (this Action) GetName() *string {
	return this.fields.GetString(ActionFieldName)
}

func (this *Action) SetName(v *string) {
	this.fields.SetString(ActionFieldName, v)
}
