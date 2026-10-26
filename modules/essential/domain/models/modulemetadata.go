package models

import (
	_ "embed"
	"slices"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	ModuleMetadataSchemaName = "essential_module_metadata"

	ModuleMetadataFieldId            = basemodel.FieldId
	ModuleMetadataFieldName          = "name"
	ModuleMetadataFieldLabel         = "label"
	ModuleMetadataFieldIsInternal    = "is_internal"
	ModuleMetadataFieldIsOrphaned    = "is_orphaned"
	ModuleMetadataFieldVersion       = "version"
	ModuleMetadataFieldHasUi         = "has_ui"
	ModuleMetadataFieldModelPrefixes = "model_prefixes"
	ModuleMetadataFieldCategory      = "category"
	ModuleMetadataFieldIsEnabled     = "is_enabled"
	ModuleMetadataFieldDeps          = "dependencies"
	ModuleMetadataFieldDepLevel      = "dep_level"
)

//go:embed module.json
var moduleMetadataSchemaJson string

func ModuleMetadataSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(moduleMetadataSchemaJson)
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

func (this ModuleMetadata) GetIsInternal() *bool {
	return this.GetFieldData().GetBool(ModuleMetadataFieldIsInternal)
}

func (this *ModuleMetadata) SetIsInternal(v *bool) {
	this.GetFieldData().SetBool(ModuleMetadataFieldIsInternal, v)
}

func (this ModuleMetadata) GetIsOrphaned() *bool {
	return this.GetFieldData().GetBool(ModuleMetadataFieldIsOrphaned)
}

func (this *ModuleMetadata) SetIsOrphaned(v *bool) {
	this.GetFieldData().SetBool(ModuleMetadataFieldIsOrphaned, v)
}

func (this ModuleMetadata) GetHasUi() *bool {
	return this.GetFieldData().GetBool(ModuleMetadataFieldHasUi)
}

func (this *ModuleMetadata) SetHasUi(v *bool) {
	this.GetFieldData().SetBool(ModuleMetadataFieldHasUi, v)
}

func (this ModuleMetadata) GetModelPrefixes() []string {
	return this.GetFieldData().GetStrings(ModuleMetadataFieldModelPrefixes)
}

func (this *ModuleMetadata) SetModelPrefixes(v []string) {
	this.GetFieldData().SetStrings(ModuleMetadataFieldModelPrefixes, v)
}

func (this ModuleMetadata) GetCategory() *modules.ModuleCategory {
	v := this.GetFieldData().GetString(ModuleMetadataFieldCategory)
	if v == nil {
		return nil
	}
	return util.ToPtr(modules.ModuleCategory(*v))
}

func (this *ModuleMetadata) SetCategory(v *modules.ModuleCategory) {
	if v == nil {
		this.GetFieldData().SetString(ModuleMetadataFieldCategory, nil)
		return
	}
	this.GetFieldData().SetString(ModuleMetadataFieldCategory, util.ToPtr(string(*v)))
}

func (this ModuleMetadata) GetIsEnabled() *bool {
	return this.GetFieldData().GetBool(ModuleMetadataFieldIsEnabled)
}

func (this *ModuleMetadata) SetIsEnabled(v *bool) {
	this.GetFieldData().SetBool(ModuleMetadataFieldIsEnabled, v)
}

func (this ModuleMetadata) GetDependencies() []string {
	return this.GetFieldData().GetStrings(ModuleMetadataFieldDeps)
}

func (this *ModuleMetadata) SetDependencies(v []string) {
	this.GetFieldData().SetStrings(ModuleMetadataFieldDeps, v)
}

func (this ModuleMetadata) GetDepLevel() *int32 {
	return this.GetFieldData().GetInt32(ModuleMetadataFieldDepLevel)
}

func (this *ModuleMetadata) SetDepLevel(v *int32) {
	this.GetFieldData().SetInt32(ModuleMetadataFieldDepLevel, v)
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

// ModifiedFields returns what the stored row must change to describe other, or nil when nothing
// differs. depLevel is other's depth among the loaded modules, which other cannot know itself.
func (this ModuleMetadata) ModifiedFields(other modules.InCodeModule, depLevel int32) *ModuleMetadata {
	modified := NewModuleMetadata()
	count := 0

	if !slices.Equal(this.GetDependencies(), other.Deps()) {
		modified.SetDependencies(other.Deps())
		count++
	}

	if this.GetDepLevel() == nil || *this.GetDepLevel() != depLevel {
		modified.SetDepLevel(&depLevel)
		count++
	}

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

	if this.GetHasUi() == nil || *this.GetHasUi() != other.HasUi() {
		modified.SetHasUi(util.ToPtr(other.HasUi()))
		count++
	}

	if !slices.Equal(this.GetModelPrefixes(), other.ModelPrefixes()) {
		modified.SetModelPrefixes(other.ModelPrefixes())
		count++
	}

	if this.GetCategory() == nil || *this.GetCategory() != other.Category() {
		modified.SetCategory(util.ToPtr(other.Category()))
		count++
	}

	if count == 0 {
		return nil
	}
	return modified
}
