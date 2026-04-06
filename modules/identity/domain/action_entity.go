package domain

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ActionSchemaName = "authorize.action"

	ActionFieldId          = basemodel.FieldId
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
		CompositeUnique(ActionFieldName, ActionFieldResourceId).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(ActionFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(ActionFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile("^[a-zA-Z0-9_-]+$"),
				})).
				RequiredForCreate().
				NoUpdate(),
		).
		Field(
			dmodel.DefineField().Name(ActionFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(ActionFieldResourceId).
				DataType(dmodel.FieldDataTypeUlid()).
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

func (this Action) GetCode() *string {
	return this.fields[ActionFieldCode].(*string)
}

func (this *Action) SetCode(v *string) {
	this.fields[ActionFieldCode] = v
}

func (this Action) GetResourceId() *model.Id {
	return this.fields[ActionFieldResourceId].(*model.Id)
}

func (this *Action) SetResourceId(v *model.Id) {
	this.fields[ActionFieldResourceId] = v
}

func (this Action) GetName() *string {
	return this.fields[ActionFieldName].(*string)
}

func (this *Action) SetName(v *string) {
	this.fields[ActionFieldName] = v
}
