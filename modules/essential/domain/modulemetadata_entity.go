package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ModuleMetadataSchemaName = "essential.module_metadata"

	ModuleMetadataFieldId         = basemodel.FieldId
	ModuleMetadataFieldName       = "name"
	ModuleMetadataFieldLabel      = "label"
	ModuleMetadataFieldIsOrphaned = "is_orphaned"
	ModuleMetadataFieldVersion    = "version"
)

func ModuleMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ModuleMetadataSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Module Metadata"}).
		TableName("essential_modules").
		CompositeUnique(ModuleMetadataFieldName).
		ShouldBuildDb().
		Field(
			basemodel.DefineFieldId(ModuleMetadataFieldId).
				Label(model.LangJson{model.LanguageCodeEnUs: "ID"}).
				UseTypeDefault().
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldName).
				Label(model.LangJson{model.LanguageCodeEnUs: "Name"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldLabel).
				Label(model.LangJson{model.LanguageCodeEnUs: "Label"}).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldIsOrphaned).
				Label(model.LangJson{model.LanguageCodeEnUs: "Is Orphaned"}).
				DataType(dmodel.FieldDataTypeBoolean()).
				RequiredForCreate().
				Default(false),
		).
		Field(
			dmodel.DefineField().
				Name(ModuleMetadataFieldVersion).
				Label(model.LangJson{model.LanguageCodeEnUs: "Version"}).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type ModuleMetadata struct {
	basemodel.DynamicModelBase
}

func NewModuleMetadata() *ModuleMetadata {
	return &ModuleMetadata{basemodel.NewDynamicModel()}
}

func NewModuleMetadataFrom(src dmodel.DynamicFields) *ModuleMetadata {
	return &ModuleMetadata{basemodel.NewDynamicModel(src)}
}

func (this ModuleMetadata) GetLabel() *model.LangJson {
	v := this.GetFieldData().GetAny(ModuleMetadataFieldLabel)
	if v == nil {
		return nil
	}
	lj := v.(model.LangJson)
	return &lj
}

func (this *ModuleMetadata) SetLabel(v *model.LangJson) {
	if v == nil {
		this.GetFieldData().SetAny(ModuleMetadataFieldLabel, nil)
		return
	}
	this.GetFieldData().SetAny(ModuleMetadataFieldLabel, *v)
}

func (this ModuleMetadata) GetName() *string {
	return this.GetFieldData().GetString(ModuleMetadataFieldName)
}

func (this *ModuleMetadata) SetName(v *string) {
	this.GetFieldData().SetString(ModuleMetadataFieldName, v)
}

func (this ModuleMetadata) GetIsOrphaned() *bool {
	return this.GetFieldData().GetBool(ModuleMetadataFieldIsOrphaned)
}

func (this *ModuleMetadata) SetIsOrphaned(v *bool) {
	this.GetFieldData().SetBool(ModuleMetadataFieldIsOrphaned, v)
}

func (this ModuleMetadata) GetVersion() *semver.SemVer {
	version := this.GetFieldData().GetString(ModuleMetadataFieldVersion)
	if version == nil {
		return nil
	}
	parsed, err := semver.ParseSemVer(*version)
	if err != nil {
		return nil
	}
	return parsed
}

func (this *ModuleMetadata) SetVersion(v *semver.SemVer) {
	if v == nil {
		this.GetFieldData().SetString(ModuleMetadataFieldVersion, nil)
		return
	}
	version := v.String()
	this.GetFieldData().SetString(ModuleMetadataFieldVersion, &version)
}

func (this ModuleMetadata) ModifiedFields(other modules.InCodeModule) *ModuleMetadata {
	modified := NewModuleMetadata()
	count := 0

	if this.GetLabel() != nil && this.GetLabel().TranslationKey() != other.LabelKey() {
		label := make(model.LangJson)
		label.SetTranslationKey(other.LabelKey())
		modified.SetLabel(util.ToPtr(label))
		count++
	}

	if this.GetVersion() != nil && *this.GetVersion() != other.Version() {
		modified.SetVersion(util.ToPtr(other.Version()))
		count++
	}

	if count == 0 {
		return nil
	}
	return modified
}
