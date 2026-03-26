package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type Group struct {
	model.ModelBase
	model.AuditableBase

	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	OrgId       *model.Id `json:"orgId"`
	ScopeRef    *model.Id `json:"scopeRef,omitempty" model:"-"`

	Org *Organization `json:"organization,omitempty" model:"-"` // TODO: Handle copy
}

func (this *Group) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.Length(0, model.MODEL_RULE_DESC_LENGTH),
		),
		model.IdPtrValidateRule(&this.OrgId, false),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

func GroupSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("identity.group").
		Label(model.LangJson{"en-US": "User Group"}).
		TableName("ident_groups").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name("name").
				Label(model.LangJson{"en-US": "Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_LONG_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name("description").
				Label(model.LangJson{"en-US": "Description"}).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		)
}
