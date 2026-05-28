package models

import (
	"regexp"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	UserPrefCodeUiListColumns = "ui.list_columns"
)

const (
	UserPreferenceSchemaName = "settings.user_preference"

	UserPrefFieldId    = "id"
	UserPrefFieldCode  = "code"
	UserPrefFieldValue = "value"
)

func UserPreferenceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(UserPreferenceSchemaName).
		Label(model.LangJson{"en-US": "User Preference"}).
		TableName("settings_user_preferences").
		CompositeUnique(UserPrefFieldCode).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(UserPrefFieldCode).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
					Regex: regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(UserPrefFieldValue).
				DataType(dmodel.FieldDataTypeJsonMap()).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder())
}

type UserPreference struct {
	basemodel.DynamicModelBase
}

func NewUserPreference() *UserPreference {
	return &UserPreference{DynamicModelBase: basemodel.NewDynamicModel()}
}

func NewUserPreferenceFrom(src dmodel.DynamicFields) *UserPreference {
	return &UserPreference{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this UserPreference) GetId() *model.Id {
	return this.GetFieldData().GetModelId(UserPrefFieldId)
}

func (this *UserPreference) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(UserPrefFieldId, v)
}

func (this UserPreference) GetCode() *string {
	return this.GetFieldData().GetString(UserPrefFieldCode)
}

func (this *UserPreference) SetCode(v *string) {
	this.GetFieldData().SetString(UserPrefFieldCode, v)
}

func (this UserPreference) GetValue() map[string]any {
	return this.GetFieldData().GetAny(UserPrefFieldValue).(map[string]any)
}

func (this *UserPreference) SetValue(v map[string]any) {
	this.GetFieldData().SetAny(UserPrefFieldValue, v)
}
