package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	cmodel "github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
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
// the default's, with the language rules chained into Create, Update and Delete and `is_active`
// written to the acting organization's activation row; it also serves the lookups other modules
// make. effectiveSettings answers which language is the default one.
func NewLanguageDomainService(
	base dynamicresource.CrudDomainService, effectiveSettings itExt.EffectiveSettingsExtService,
	activation Activations,
) it.LanguageDomainService {
	return &LanguageDomainServiceImpl{CrudDomainService: base, effectiveSettings: effectiveSettings, activation: activation}
}

type LanguageDomainServiceImpl struct {
	dynamicresource.CrudDomainService
	effectiveSettings itExt.EffectiveSettingsExtService
	activation        Activations
}

func (this *LanguageDomainServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand, options ...dynamicresource.CreateOptions,
) (*dynamicresource.CreateResult, error) {
	change := popActiveChange(cmd, true)
	opts := safe.GetOptional(options, dynamicresource.CreateOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateWriteOf(change))
	var result *dynamicresource.CreateResult
	err := dynamicresource.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		created, err := this.CrudDomainService.Create(tranxCtx, cmd, opts)
		result = created
		if err != nil || created == nil || created.ClientErrors.Count() > 0 {
			return err
		}
		return applyActiveChange(tranxCtx, this.activation, created.Data[models.LanguageFieldId], change)
	})
	return result, errors.Wrap(err, "create language")
}

func (this *LanguageDomainServiceImpl) Update(
	ctx corectx.Context, cmd dynamicresource.UpdateCommand, options ...dynamicresource.UpdateOptions,
) (*dynamicresource.MutateResult, error) {
	change := popActiveChange(cmd, false)
	opts := safe.GetOptional(options, dynamicresource.UpdateOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateWriteOf(change))
	var result *dynamicresource.MutateResult
	err := dynamicresource.InTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		updated, err := this.CrudDomainService.Update(tranxCtx, cmd, opts)
		result = updated
		if err != nil || updated == nil || updated.ClientErrors.Count() > 0 {
			return err
		}
		return applyActiveChange(tranxCtx, this.activation, cmd[models.LanguageFieldId], change)
	})
	return result, errors.Wrap(err, "update language")
}

func (this *LanguageDomainServiceImpl) Delete(
	ctx corectx.Context, cmd dynamicresource.DeleteCommand, options ...dynamicresource.DeleteOptions,
) (*dynamicresource.MutateResult, error) {
	opts := safe.GetOptional(options, dynamicresource.DeleteOptions{})
	opts.ValidateExtra = dynamicresource.ChainValidateExtra(opts.ValidateExtra, this.validateDelete)
	return this.CrudDomainService.Delete(ctx, cmd, opts)
}

func (this *LanguageDomainServiceImpl) validateWriteOf(change activeChange) dynamicresource.ValidateExtraFn {
	return func(
		ctx corectx.Context, inputModel *dynamicresource.DynamicEntity, foundModel *dynamicresource.DynamicEntity,
		vErrs *ft.ClientErrors,
	) error {
		return this.validateWrite(ctx, inputModel, foundModel, vErrs, change)
	}
}

func (this *LanguageDomainServiceImpl) validateWrite(
	ctx corectx.Context, inputModel *dynamicresource.DynamicEntity, foundModel *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors, change activeChange,
) error {
	view := newLanguageView(inputModel, foundModel)
	assertLanguageFields(view, vErrs)
	if vErrs.Count() > 0 || foundModel == nil || !change.deactivates() {
		return nil
	}
	return this.assertCanLeaveActiveSet(ctx, view, vErrs)
}

// validateDelete receives the stored row as foundModel; a language the acting org does not use
// may always go.
func (this *LanguageDomainServiceImpl) validateDelete(
	ctx corectx.Context, _ *dynamicresource.DynamicEntity, foundModel *dynamicresource.DynamicEntity,
	vErrs *ft.ClientErrors,
) error {
	return this.assertCanLeaveActiveSet(ctx, newLanguageView(nil, foundModel), vErrs)
}

// assertCanLeaveActiveSet applies, for the acting organization, the rules a language leaving its
// active set must pass. Without an acting org there is no active set to protect.
func (this *LanguageDomainServiceImpl) assertCanLeaveActiveSet(
	ctx corectx.Context, view languageView, vErrs *ft.ClientErrors,
) error {
	orgId, ok := dyn.ActingOrgOf(ctx)
	if !ok {
		return nil
	}
	selfId := cmodel.Id(view.str(models.LanguageFieldId))
	active, err := this.activation.IsActive(ctx, orgId, selfId)
	if err != nil || !active {
		return err
	}
	if err := this.assertNotTheDefaultLanguage(ctx, view, vErrs); err != nil || vErrs.Count() > 0 {
		return err
	}
	return this.assertKeepsAnActiveLanguage(ctx, orgId, selfId, vErrs)
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

// ListEnabledLanguages lists the languages the acting organization has active -- else the only
// organization the caller belongs to -- or, with no organization at all (the sign-in page), the
// languages marked public_use. Archived languages are never listed.
func (this *LanguageDomainServiceImpl) ListEnabledLanguages(
	ctx corectx.Context, _ it.ListEnabledLanguagesQuery,
) (*it.ListEnabledLanguagesResult, error) {
	graph := dmodel.NewSearchGraph()
	if orgId, ok := resolveListingOrg(ctx); ok {
		ids, err := this.activation.ActiveIds(ctx, orgId)
		if err != nil {
			return nil, errors.Wrap(err, "list enabled languages")
		}
		if len(ids) == 0 {
			return &it.ListEnabledLanguagesResult{Data: []models.Language{}, HasData: true}, nil
		}
		graph.NewCondition(models.LanguageFieldId, dmodel.In, idsAsAny(ids)...)
	} else {
		graph.NewCondition(models.LanguageFieldPublicUse, dmodel.Equals, true)
	}
	graph.OrderBy(models.LanguageFieldName)
	found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph, Page: 0, Size: enabledLanguagesLimit, IncludeArchived: util.ToPtr(false),
	})
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
