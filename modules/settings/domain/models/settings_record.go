package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SettingsRecordSchemaName = "settings_record"

	SettingsRecordFieldId        = basemodel.FieldId
	SettingsRecordFieldCreatedAt = basemodel.FieldCreatedAt
	SettingsRecordFieldUpdatedAt = basemodel.FieldUpdatedAt

	SettingsRecordFieldSchemaId      = "schema_id"
	SettingsRecordFieldModuleKey     = "module_key"
	SettingsRecordFieldLevel         = "level"
	SettingsRecordFieldOwnerType     = "owner_type"
	SettingsRecordFieldOwnerId       = "owner_id"
	SettingsRecordFieldName          = "name"
	SettingsRecordFieldValue         = "value"
	SettingsRecordFieldAllowOverride = "allow_override"
)

// ValueEnvelopeKey is the single key of the `value` column's object. The envelope exists because
// json_map has no array form, so a scalar and a list both need an object around them to share one
// column.
const ValueEnvelopeKey = "value"

//go:embed settings_record.json
var settingsRecordSchemaJson string

func SettingsRecordSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(settingsRecordSchemaJson)
}

// SettingsRecord is one setting item's stored value for one owner.
type SettingsRecord struct {
	basemodel.DynamicModelBase
}

func NewSettingsRecord() *SettingsRecord {
	return &SettingsRecord{DynamicModelBase: basemodel.NewDynamicModel()}
}

func NewSettingsRecordFrom(src dmodel.DynamicFields) *SettingsRecord {
	return &SettingsRecord{DynamicModelBase: basemodel.NewDynamicModel(src)}
}

func (this SettingsRecord) GetSchemaId() *model.Id {
	return this.GetFieldData().GetModelId(SettingsRecordFieldSchemaId)
}

func (this *SettingsRecord) SetSchemaId(schemaId *model.Id) {
	this.GetFieldData().SetModelId(SettingsRecordFieldSchemaId, schemaId)
}

func (this SettingsRecord) GetModuleKey() *string {
	return this.GetFieldData().GetString(SettingsRecordFieldModuleKey)
}

func (this *SettingsRecord) SetModuleKey(moduleKey *string) {
	this.GetFieldData().SetString(SettingsRecordFieldModuleKey, moduleKey)
}

func (this SettingsRecord) GetLevel() *string {
	return this.GetFieldData().GetString(SettingsRecordFieldLevel)
}

func (this *SettingsRecord) SetLevel(level *string) {
	this.GetFieldData().SetString(SettingsRecordFieldLevel, level)
}

func (this SettingsRecord) GetOwnerType() *string {
	return this.GetFieldData().GetString(SettingsRecordFieldOwnerType)
}

func (this *SettingsRecord) SetOwnerType(ownerType *string) {
	this.GetFieldData().SetString(SettingsRecordFieldOwnerType, ownerType)
}

func (this SettingsRecord) GetOwnerId() *model.Id {
	return this.GetFieldData().GetModelId(SettingsRecordFieldOwnerId)
}

func (this *SettingsRecord) SetOwnerId(ownerId *model.Id) {
	this.GetFieldData().SetModelId(SettingsRecordFieldOwnerId, ownerId)
}

func (this SettingsRecord) GetName() *string {
	return this.GetFieldData().GetString(SettingsRecordFieldName)
}

func (this *SettingsRecord) SetName(name *string) {
	this.GetFieldData().SetString(SettingsRecordFieldName, name)
}

// GetValue unwraps the stored `{"value": ...}` envelope and reports whether a value was present.
// A null column, a non-object column and an object without the envelope key are all reported as
// absent rather than panicking, since all three are reachable from the database.
func (this SettingsRecord) GetValue() (any, bool) {
	raw := this.GetFieldData().GetAny(SettingsRecordFieldValue)
	if raw == nil {
		return nil, false
	}
	envelope, ok := raw.(map[string]any)
	if !ok {
		return nil, false
	}
	val, ok := envelope[ValueEnvelopeKey]
	return val, ok
}

// SetValue stores val inside the envelope, replacing whatever the column held.
func (this *SettingsRecord) SetValue(val any) {
	this.GetFieldData().SetAny(SettingsRecordFieldValue, map[string]any{ValueEnvelopeKey: val})
}

// GetAllowOverride reports whether an owner below the tenant may keep their own value.
//
// Nil when no tenant admin has ruled on this setting. Callers must read that as overridable: the
// restrictive reading would silently lock every setting nobody has considered, which is the same
// policy the schema-metadata reader followed before the flag moved onto the row.
func (this SettingsRecord) GetAllowOverride() *bool {
	return this.GetFieldData().GetBool(SettingsRecordFieldAllowOverride)
}

func (this *SettingsRecord) SetAllowOverride(allow *bool) {
	this.GetFieldData().SetBool(SettingsRecordFieldAllowOverride, allow)
}
