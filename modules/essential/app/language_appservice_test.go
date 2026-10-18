package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	langjson "github.com/sky-as-code/nikki-erp/modules/essential/infra/langJson"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

const testModule = "vendingmachine"

func newLanguageService(t *testing.T) (it.LanguageAppService, corectx.Context) {
	t.Helper()
	// The language JSON methods never reach the domain service.
	return app.NewLanguageApplicationServiceForTest(), corectx.NewRequestContext(context.Background())
}

func getJson(
	t *testing.T, svc it.LanguageAppService, ctx corectx.Context, lang string, module string,
) *it.GetLanguageJsonResult {
	t.Helper()
	result, err := svc.GetLanguageJson(ctx, it.GetLanguageJsonQuery{LanguageCode: lang, ModuleName: module})
	require.NoError(t, err)
	return result
}

func TestRegisteredPackIsServedUnderModuleName(t *testing.T) {
	svc, ctx := newLanguageService(t)
	require.NoError(t, svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{
		"en-US": {"kiosk.title": "Kiosk"},
		"vi-VN": {"kiosk.title": "Máy bán hàng"},
	}))

	result := getJson(t, svc, ctx, "vi-VN", testModule)
	require.True(t, result.HasData)
	assert.Equal(t, "Máy bán hàng", result.Data["kiosk.title"])
}

func TestLanguageOnlyRegisteredByAModuleIsKnown(t *testing.T) {
	svc, ctx := newLanguageService(t)
	require.NoError(t, svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{"fr-FR": {"a": "b"}}))

	assert.True(t, getJson(t, svc, ctx, "fr-FR", testModule).HasData)
}

func TestSecondRegistrationOfSamePackFails(t *testing.T) {
	svc, ctx := newLanguageService(t)
	require.NoError(t, svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{"en-US": {"k": "first"}}))

	err := svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{"en-US": {"k": "second"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), testModule)
	assert.Equal(t, "first", getJson(t, svc, ctx, "en-US", testModule).Data["k"])
}

func TestPackOverlappingAnEmbeddedOneFails(t *testing.T) {
	svc, ctx := newLanguageService(t)

	err := svc.RegisterModuleLanguage("iam", model.LanguageTranslation{
		"fr-FR": {"k": "v"},
		"en-US": {"k": "v"},
	})
	require.Error(t, err)
	// Rejected as a whole: the non-conflicting language must not have been stored either.
	assert.False(t, getJson(t, svc, ctx, "fr-FR", "iam").HasData)
}

func TestRegistrationRequiresModuleName(t *testing.T) {
	svc, _ := newLanguageService(t)

	assert.Error(t, svc.RegisterModuleLanguage("", model.LanguageTranslation{"en-US": {}}))
}

func TestRegistrationRejectsInvalidLanguageCode(t *testing.T) {
	svc, ctx := newLanguageService(t)

	err := svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{"english": {"k": "v"}})
	require.Error(t, err)
	assert.False(t, getJson(t, svc, ctx, "english", testModule).HasData)
}

func TestUnknownLanguageOrModuleIsNotFound(t *testing.T) {
	svc, ctx := newLanguageService(t)

	unknownLang := getJson(t, svc, ctx, "xx-XX", "iam")
	assert.False(t, unknownLang.HasData)
	assert.True(t, unknownLang.ClientErrors.Has("language_code"))

	unknownModule := getJson(t, svc, ctx, "en-US", "no_such_module")
	assert.False(t, unknownModule.HasData)
	assert.True(t, unknownModule.ClientErrors.Has("module_name"))
}

func TestCommonIncludesRegisteredModuleLabelsWithoutMutatingEmbeddedPack(t *testing.T) {
	svc, ctx := newLanguageService(t)
	const labelKey = langjson.ModuleLabelKeyPrefix + ".vendingmachine_test"
	require.NoError(t, svc.RegisterModuleLanguage(testModule, model.LanguageTranslation{
		"en-US": {labelKey: "Vending Machine", "kiosk.title": "Kiosk"},
	}))

	common := getJson(t, svc, ctx, "en-US", langjson.CommonModuleName)
	require.True(t, common.HasData)
	assert.Equal(t, "Vending Machine", common.Data[labelKey])
	assert.NotContains(t, common.Data, "kiosk.title")

	embeddedCommon := langjson.EmbeddedPack[langjson.CommonModuleName]["en-US"]
	assert.NotContains(t, embeddedCommon, labelKey)
}
