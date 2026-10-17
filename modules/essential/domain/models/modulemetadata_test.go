package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
)

type categoryTestModule struct {
	category modules.ModuleCategory
}

func (this *categoryTestModule) Deps() []string                   { return nil }
func (this *categoryTestModule) LabelKey() string                 { return "test.moduleLabel" }
func (this *categoryTestModule) Name() string                     { return "test" }
func (this *categoryTestModule) Init() error                      { return nil }
func (this *categoryTestModule) IsInternal() bool                 { return false }
func (this *categoryTestModule) ModelPrefix() string              { return "test" }
func (this *categoryTestModule) HasUi() bool                      { return true }
func (this *categoryTestModule) Category() modules.ModuleCategory { return this.category }
func (this *categoryTestModule) Version() semver.SemVer           { return *semver.MustParseSemVer("v1.0.0") }

// syncedMetadata returns the row the startup sync would have written for mod, so that only the
// field under test differs from it.
func syncedMetadata(mod modules.InCodeModule, category *modules.ModuleCategory) *ModuleMetadata {
	label := make(model.LangJson)
	label.SetTranslationKey(mod.LabelKey())

	meta := NewModuleMetadata()
	meta.SetLabel(&label)
	meta.SetVersion(util.ToPtr(mod.Version()))
	meta.SetHasUi(util.ToPtr(mod.HasUi()))
	meta.SetModelPrefix(util.ToPtr(mod.ModelPrefix()))
	meta.SetCategory(category)
	meta.SetIsEnabled(util.ToPtr(true))
	return meta
}

func TestModifiedFieldsUnchangedCategory(t *testing.T) {
	mod := &categoryTestModule{category: modules.ModuleCategoryCoremart}
	meta := syncedMetadata(mod, util.ToPtr(modules.ModuleCategoryCoremart))

	assert.Nil(t, meta.ModifiedFields(mod))
}

func TestModifiedFieldsChangedCategory(t *testing.T) {
	mod := &categoryTestModule{category: modules.ModuleCategoryCoreparking}
	meta := syncedMetadata(mod, util.ToPtr(modules.ModuleCategoryFoundation))

	modified := meta.ModifiedFields(mod)
	require.NotNil(t, modified)
	assert.Equal(t, util.ToPtr(modules.ModuleCategoryCoreparking), modified.GetCategory())
	assert.Nil(t, modified.GetIsEnabled(), "is_enabled is never part of the sync diff")
}

// A row written before the category column existed holds NULL; the sync backfills it.
func TestModifiedFieldsBackfillsMissingCategory(t *testing.T) {
	mod := &categoryTestModule{category: modules.ModuleCategoryFoundation}
	meta := syncedMetadata(mod, nil)

	modified := meta.ModifiedFields(mod)
	require.NotNil(t, modified)
	assert.Equal(t, util.ToPtr(modules.ModuleCategoryFoundation), modified.GetCategory())
}
