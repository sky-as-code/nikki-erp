package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
)

type languageStubRepo struct {
	dynamicresource.CrudRepository
	enabled []dmodel.DynamicFields
}

func (this *languageStubRepo) Search(corectx.Context, dyn.RepoSearchParam) (*dynamicresource.SearchResult, error) {
	data := dyn.PagedResultData[dmodel.DynamicFields]{Items: this.enabled}
	return &dynamicresource.SearchResult{Data: data, HasData: true}, nil
}

type languageStubDomain struct {
	dynamicresource.CrudDomainService
	repo *languageStubRepo
}

func (this *languageStubDomain) Repository() dynamicresource.CrudRepository { return this.repo }

type stubEffectiveSettings struct{ systemLanguage string }

func (this stubEffectiveSettings) GetEffectiveSettings(
	corectx.Context, itExt.GetEffectiveSettingsQuery,
) (*itExt.GetEffectiveSettingsResult, error) {
	result := &itExt.GetEffectiveSettingsResult{HasData: true}
	result.Data.Values = map[string]any{"essential.system_language": this.systemLanguage}
	return result, nil
}

func newLanguageSvc(systemLanguage string, enabledIds ...string) *LanguageDomainServiceImpl {
	repo := &languageStubRepo{}
	for _, id := range enabledIds {
		repo.enabled = append(repo.enabled, dmodel.DynamicFields{"id": id})
	}
	return NewLanguageDomainService(
		&languageStubDomain{repo: repo}, stubEffectiveSettings{systemLanguage: systemLanguage},
	).(*LanguageDomainServiceImpl)
}

func userContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{UserId: "01JQZ0X0000000000000000001"})
	return ctx
}

func entity(fields dmodel.DynamicFields) *dynamicresource.DynamicEntity {
	return dynamicresource.NewDynamicEntityFrom(fields)
}

func storedLanguage(id, code string, enabled bool) *dynamicresource.DynamicEntity {
	return entity(dmodel.DynamicFields{
		"id": id, "locale_code": code, "is_enabled": enabled,
		"decimal_separator": ".", "thousands_separator": ",",
	})
}

func requireOneError(t *testing.T, vErrs *ft.ClientErrors, field string, key string) {
	t.Helper()
	require.Equal(t, 1, vErrs.Count(), "%v", vErrs)
	item := (*vErrs)[0]
	assert.Equal(t, field, item.Field)
	assert.Equal(t, key, item.Key)
}

func TestLanguageFormatRejectsOdooDisallowedDirectives(t *testing.T) {
	for _, format := range []string{"%e/%m/%Y", "%D", "%H:%M:%S %P", "%T", "%+"} {
		vErrs := ft.NewClientErrors()
		view := newLanguageView(entity(dmodel.DynamicFields{"date_format": format}), storedLanguage("a", "en-US", true))

		assertLanguageFields(view, vErrs)

		requireOneError(t, vErrs, "date_format", "essential:err_lang_format_directive")
	}
}

func TestLanguageFormatAcceptsPortableDirectives(t *testing.T) {
	for _, format := range []string{"%m/%d/%Y", "%d/%m/%y", "%H:%M:%S", "%I:%M %p", "%A %d %B %Y"} {
		vErrs := ft.NewClientErrors()
		view := newLanguageView(entity(dmodel.DynamicFields{"time_format": format}), storedLanguage("a", "en-US", true))

		assertLanguageFields(view, vErrs)

		assert.Zero(t, vErrs.Count(), "format %q must be allowed", format)
	}
}

func TestLanguageGroupingFollowsOdooSyntax(t *testing.T) {
	for _, grouping := range []string{"[3]", "[3,0]", "[3,2,-1]", "[]", " [3, 0] "} {
		assert.True(t, isValidGrouping(grouping), grouping)
	}
	for _, grouping := range []string{"3", "[3.5]", "[\"3\"]", "[3,", "{}", "", "[3] [2]", "null"} {
		assert.False(t, isValidGrouping(grouping), grouping)
	}
}

func TestLanguageSeparatorsMustDiffer(t *testing.T) {
	vErrs := ft.NewClientErrors()
	view := newLanguageView(entity(dmodel.DynamicFields{"thousands_separator": "."}), storedLanguage("a", "en-US", true))

	assertLanguageFields(view, vErrs)

	requireOneError(t, vErrs, "thousands_separator", "essential:err_lang_separators_equal")
}

func TestLanguageEmptyThousandsSeparatorIsAllowed(t *testing.T) {
	vErrs := ft.NewClientErrors()
	view := newLanguageView(entity(dmodel.DynamicFields{"thousands_separator": ""}), storedLanguage("a", "en-US", true))

	assertLanguageFields(view, vErrs)

	assert.Zero(t, vErrs.Count())
}

func TestDisablingTheLastEnabledLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{"is_enabled": false}),
		storedLanguage("a", "en-US", true), vErrs)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsEnabled, "essential:err_lang_last_enabled")
}

func TestDisablingWhileAnotherStaysEnabledIsAllowed(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{"is_enabled": false}),
		storedLanguage("a", "en-US", true), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestDeletingTheLastEnabledLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a")
	vErrs := ft.NewClientErrors()

	err := svc.validateDelete(userContext(), nil, storedLanguage("a", "en-US", true), vErrs)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsEnabled, "essential:err_lang_last_enabled")
}

func TestDeletingADisabledLanguageSkipsTheEnabledRules(t *testing.T) {
	svc := newLanguageSvc("en-US")
	vErrs := ft.NewClientErrors()

	err := svc.validateDelete(userContext(), nil, storedLanguage("a", "en-US", false), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestDisablingTheDefaultLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{"is_enabled": false}),
		storedLanguage("b", "vi-VN", true), vErrs)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsEnabled, "essential:err_lang_default_disabled")
}

// With no system_language stored the application default is what users fall back to, so it is
// the default language for this rule.
func TestApplicationDefaultCountsWhenNoSystemLanguageIsSet(t *testing.T) {
	svc := newLanguageSvc("", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{"is_enabled": false}),
		storedLanguage("a", "en-US", true), vErrs)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsEnabled, "essential:err_lang_default_disabled")
}
