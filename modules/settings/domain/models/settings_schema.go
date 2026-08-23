package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SettingsSchemaSchemaName = "settings_schema"

	SettingsSchemaFieldId        = basemodel.FieldId
	SettingsSchemaFieldCreatedAt = basemodel.FieldCreatedAt
	SettingsSchemaFieldUpdatedAt = basemodel.FieldUpdatedAt

	SettingsSchemaFieldModuleKey = "module_key"
	SettingsSchemaFieldLevel     = "level"
	SettingsSchemaFieldSchema    = "schema"
)

//go:embed settings_schema.json
var settingsSchemaSchemaJson string

func SettingsSchemaSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(settingsSchemaSchemaJson)
}

// SettingsSchema is one module's setting definitions for one level.
type SettingsSchema struct {
	basemodel.DynamicModelBase
}

func NewSettingsSchema() *SettingsSchema {
	return &SettingsSchema{DynamicModelBase: basemodel.NewDynamicModel()}
}

func NewSettingsSchemaFrom(src dmodel.DynamicFields) *SettingsSchema {
	return &SettingsSchema{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this SettingsSchema) GetModuleKey() *string {
	return this.GetFieldData().GetString(SettingsSchemaFieldModuleKey)
}

func (this *SettingsSchema) SetModuleKey(moduleKey *string) {
	this.GetFieldData().SetString(SettingsSchemaFieldModuleKey, moduleKey)
}

func (this SettingsSchema) GetLevel() *string {
	return this.GetFieldData().GetString(SettingsSchemaFieldLevel)
}

func (this *SettingsSchema) SetLevel(level *string) {
	this.GetFieldData().SetString(SettingsSchemaFieldLevel, level)
}

// GetSchema returns the stored model-schema document, or nil when the column is null or holds
// something other than an object. Checked rather than asserted: a null column is reachable from
// the database, and an unchecked assertion would panic the request instead of failing it.
func (this SettingsSchema) GetSchema() map[string]any {
	raw := this.GetFieldData().GetAny(SettingsSchemaFieldSchema)
	if raw == nil {
		return nil
	}
	schema, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return schema
}

func (this *SettingsSchema) SetSchema(schema map[string]any) {
	this.GetFieldData().SetAny(SettingsSchemaFieldSchema, schema)
}
