package services

import (
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
)

type languageStubDomain struct {
	dynamicresource.CrudDomainService
}

// stubActivations is one organization's active set.
type stubActivations struct{ active []model.Id }

func (this *stubActivations) SetActive(corectx.Context, model.Id, model.Id, bool) error { return nil }

func (this *stubActivations) ActiveIds(corectx.Context, model.Id) ([]model.Id, error) {
	return this.active, nil
}

func (this *stubActivations) IsActive(_ corectx.Context, _ model.Id, id model.Id) (bool, error) {
	return slices.Contains(this.active, id), nil
}

type stubEffectiveSettings struct{ systemLanguage string }

func (this stubEffectiveSettings) GetEffectiveSettings(
	corectx.Context, itExt.GetEffectiveSettingsQuery,
) (*itExt.GetEffectiveSettingsResult, error) {
	result := &itExt.GetEffectiveSettingsResult{HasData: true}
	result.Data.Values = map[string]any{"essential.system_language": this.systemLanguage}
	return result, nil
}

func newLanguageSvc(systemLanguage string, activeIds ...string) *LanguageDomainServiceImpl {
	activations := &stubActivations{}
	for _, id := range activeIds {
		activations.active = append(activations.active, model.Id(id))
	}
	return NewLanguageDomainService(
		&languageStubDomain{}, stubEffectiveSettings{systemLanguage: systemLanguage}, activations,
	).(*LanguageDomainServiceImpl)
}

// userContext acts in one organization: the language rules answer for its active set.
func userContext() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{UserId: "01JQZ0X0000000000000000001"})
	dyn.WithActingOrg(ctx, "01JQZ0X0000000000000000002")
	return ctx
}

func entity(fields dmodel.DynamicFields) *dynamicresource.DynamicEntity {
	return dynamicresource.NewDynamicEntityFrom(fields)
}

func storedLanguage(id, code string) *dynamicresource.DynamicEntity {
	return entity(dmodel.DynamicFields{
		"id": id, "locale_code": code, "decimal_separator": ".", "thousands_separator": ",",
	})
}

var deactivate = activeChange{sent: true, active: false}

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
		view := newLanguageView(entity(dmodel.DynamicFields{"date_format": format}), storedLanguage("a", "en-US"))

		assertLanguageFields(view, vErrs)

		requireOneError(t, vErrs, "date_format", "essential:err_lang_format_directive")
	}
}

func TestLanguageFormatAcceptsPortableDirectives(t *testing.T) {
	for _, format := range []string{"%m/%d/%Y", "%d/%m/%y", "%H:%M:%S", "%I:%M %p", "%A %d %B %Y"} {
		vErrs := ft.NewClientErrors()
		view := newLanguageView(entity(dmodel.DynamicFields{"time_format": format}), storedLanguage("a", "en-US"))

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
	view := newLanguageView(entity(dmodel.DynamicFields{"thousands_separator": "."}), storedLanguage("a", "en-US"))

	assertLanguageFields(view, vErrs)

	requireOneError(t, vErrs, "thousands_separator", "essential:err_lang_separators_equal")
}

func TestLanguageEmptyThousandsSeparatorIsAllowed(t *testing.T) {
	vErrs := ft.NewClientErrors()
	view := newLanguageView(entity(dmodel.DynamicFields{"thousands_separator": ""}), storedLanguage("a", "en-US"))

	assertLanguageFields(view, vErrs)

	assert.Zero(t, vErrs.Count())
}

func TestDeactivatingTheOrgsLastActiveLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{}), storedLanguage("a", "en-US"), vErrs, deactivate)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsActive, "essential:err_lang_last_enabled")
}

func TestDeactivatingWhileAnotherStaysActiveIsAllowed(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{}), storedLanguage("a", "en-US"), vErrs, deactivate)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestDeletingTheOrgsLastActiveLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a")
	vErrs := ft.NewClientErrors()

	err := svc.validateDelete(userContext(), nil, storedLanguage("a", "en-US"), vErrs)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsActive, "essential:err_lang_last_enabled")
}

func TestDeletingALanguageTheOrgDoesNotUseSkipsTheRules(t *testing.T) {
	svc := newLanguageSvc("en-US")
	vErrs := ft.NewClientErrors()

	err := svc.validateDelete(userContext(), nil, storedLanguage("a", "en-US"), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestDeactivatingTheDefaultLanguageIsRefused(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{}), storedLanguage("b", "vi-VN"), vErrs, deactivate)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsActive, "essential:err_lang_default_disabled")
}

// With no system_language stored the application default is what users fall back to, so it is
// the default language for this rule.
func TestApplicationDefaultCountsWhenNoSystemLanguageIsSet(t *testing.T) {
	svc := newLanguageSvc("", "a", "b")
	vErrs := ft.NewClientErrors()

	err := svc.validateWrite(userContext(), entity(dmodel.DynamicFields{}), storedLanguage("a", "en-US"), vErrs, deactivate)

	require.NoError(t, err)
	requireOneError(t, vErrs, models.LanguageFieldIsActive, "essential:err_lang_default_disabled")
}

// Without an acting org there is no active set to protect.
func TestLanguageRulesNeedAnActingOrg(t *testing.T) {
	svc := newLanguageSvc("vi-VN", "a")
	vErrs := ft.NewClientErrors()
	ctx := corectx.NewRequestContext(context.Background())

	err := svc.validateDelete(ctx, nil, storedLanguage("a", "en-US"), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

func TestPopActiveChange(t *testing.T) {
	cmd := dmodel.DynamicFields{"is_active": false, "name": "x"}
	assert.Equal(t, activeChange{sent: true, active: false}, popActiveChange(cmd, true))
	assert.NotContains(t, cmd, "is_active")
	assert.Equal(t, activeChange{sent: true, active: true}, popActiveChange(dmodel.DynamicFields{}, true))
	assert.Equal(t, activeChange{}, popActiveChange(dmodel.DynamicFields{}, false))
}
