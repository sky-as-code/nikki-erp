package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	c "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewLanguageDomainService is handed the composable default by the language onion. The CRUD is
// the default's, with the language rules chained into Create, Update and Delete; it also serves
// the lookups other modules make. effectiveSettings answers which language is the default one.
func NewLanguageDomainService(
	base dynamicresource.CrudDomainService, effectiveSettings itExt.EffectiveSettingsExtService,
) it.LanguageDomainService {
	return &LanguageDomainServiceImpl{CrudDomainService: base, effectiveSettings: effectiveSettings}
}

type LanguageDomainServiceImpl struct {
	dynamicresource.CrudDomainService
	effectiveSettings itExt.EffectiveSettingsExtService
}

func (this *LanguageDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.CreateOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateWrite)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *LanguageDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.UpdateOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateWrite)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

func (this *LanguageDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.DeleteOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateDelete)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func (this *LanguageDomainServiceImpl) validateWrite(
	ctx corectx.Context, inputModel *dynamicresource.DynamicEntity, foundModel *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	view := newLanguageView(inputModel, foundModel)
	assertLanguageFields(view, vErrs)
	if vErrs.Count() > 0 || !view.disables() {
		return nil
	}
	return this.assertCanLeaveEnabledSet(ctx, view, vErrs)
}

// validateDelete receives the stored row as foundModel; a disabled language may always go.
func (this *LanguageDomainServiceImpl) validateDelete(
	ctx corectx.Context, _ *dynamicresource.DynamicEntity, foundModel *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	view := newLanguageView(nil, foundModel)
	if !view.wasEnabled() {
		return nil
	}
	return this.assertCanLeaveEnabledSet(ctx, view, vErrs)
}

func (this *LanguageDomainServiceImpl) assertCanLeaveEnabledSet(
	ctx corectx.Context, view languageView, vErrs *ft.ClientErrors,
) error {
	if err := this.assertNotTheDefaultLanguage(ctx, view, vErrs); err != nil || vErrs.Count() > 0 {
		return err
	}
	return this.assertKeepsAnEnabledLanguage(ctx, view, vErrs)
}

// defaultLocaleCode is the acting organization's effective system_language, or the application
// default when none is set -- the same fallback the locale resolver walks. Without an acting user
// there is no organization to read, so the application default answers.
func (this *LanguageDomainServiceImpl) defaultLocaleCode(ctx corectx.Context) (string, error) {
	if this.effectiveSettings == nil || ctx.GetPermissions().UserId == "" {
		return cmodel.DefaultLanguageCode, nil
	}
	result, err := this.effectiveSettings.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{c.EssentialModuleName},
	})
	if err != nil {
		return "", errors.Wrap(err, "read the default language")
	}
	if result == nil || !result.HasData {
		return cmodel.DefaultLanguageCode, nil
	}
	key := c.EssentialModuleName + "." + settings.OrgSettingSystemLanguage
	if code, ok := result.Data.Values[key].(string); ok && code != "" {
		return code, nil
	}
	return cmodel.DefaultLanguageCode, nil
}

// GetLanguageByLocaleCode resolves the code a setting or an i18n layer names a language with. The
// locale_code column is unique, so a match is exactly one row.
func (this *LanguageDomainServiceImpl) GetLanguageByLocaleCode(
	ctx corectx.Context, query it.GetLanguageByLocaleCodeQuery,
) (*it.GetLanguageByLocaleCodeResult, error) {
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.LanguageFieldLocaleCode: query.LocaleCode},
	})
	if err != nil {
		return nil, errors.Wrap(err, "get language by locale code")
	}
	if !found.HasData {
		// Absence is a legitimate answer: the caller decides what an unknown language means.
		return &it.GetLanguageByLocaleCodeResult{HasData: false}, nil
	}
	return &it.GetLanguageByLocaleCodeResult{Data: *models.NewLanguageFrom(found.Data), HasData: true}, nil
}

// enabledLanguagesLimit bounds the read; a catalogue of locales the product ships translations for
// stays far below it.
const enabledLanguagesLimit = 500

func (this *LanguageDomainServiceImpl) ListEnabledLanguages(
	ctx corectx.Context, query it.ListEnabledLanguagesQuery,
) (*it.ListEnabledLanguagesResult, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(models.LanguageFieldIsEnabled, dmodel.Equals, true).
		OrderBy(models.LanguageFieldName)
	found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: enabledLanguagesLimit})
	if err != nil {
		return nil, errors.Wrap(err, "list enabled languages")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "list enabled languages")
	}
	languages := []models.Language{}
	if found.HasData {
		for _, row := range found.Data.Items {
			languages = append(languages, *models.NewLanguageFrom(row))
		}
	}
	return &it.ListEnabledLanguagesResult{Data: languages, HasData: true}, nil
}
