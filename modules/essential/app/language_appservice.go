package app

import (
	"sync"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	langjson "github.com/sky-as-code/nikki-erp/modules/essential/infra/langJson"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

func NewLanguageApplicationServiceImpl(languageSvc it.LanguageDomainService) it.LanguageAppService {
	return &LanguageApplicationServiceImpl{
		languageSvc: languageSvc,
		registered:  make(model.LanguagePack),
	}
}

type LanguageApplicationServiceImpl struct {
	languageSvc it.LanguageDomainService
	// registered holds the translations modules ship themselves; langjson.EmbeddedPack stays the
	// baseline.
	registered model.LanguagePack
	mutex      sync.RWMutex
}

func (this *LanguageApplicationServiceImpl) CreateLanguage(ctx corectx.Context, cmd it.CreateLanguageCommand) (*it.CreateLanguageResult, error) {
	return this.languageSvc.CreateLanguage(ctx, cmd)
}

func (this *LanguageApplicationServiceImpl) DeleteLanguage(ctx corectx.Context, cmd it.DeleteLanguageCommand) (*it.DeleteLanguageResult, error) {
	return this.languageSvc.DeleteLanguage(ctx, cmd)
}

func (this *LanguageApplicationServiceImpl) LanguageExists(ctx corectx.Context, query it.LanguageExistsQuery) (*it.LanguageExistsResult, error) {
	return this.languageSvc.LanguageExists(ctx, query)
}

func (this *LanguageApplicationServiceImpl) GetLanguage(ctx corectx.Context, query it.GetLanguageQuery) (*it.GetLanguageResult, error) {
	return this.languageSvc.GetLanguage(ctx, query)
}

func (this *LanguageApplicationServiceImpl) GetLanguageByIsoCode(
	ctx corectx.Context, query it.GetLanguageByIsoCodeQuery,
) (*it.GetLanguageByIsoCodeResult, error) {
	return this.languageSvc.GetLanguageByIsoCode(ctx, query)
}

func (this *LanguageApplicationServiceImpl) GetLanguageJson(ctx corectx.Context, query it.GetLanguageJsonQuery) (*it.GetLanguageJsonResult, error) {
	_ = ctx
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	langCode := query.LanguageCode
	if !langjson.HasLanguage(this.registered, langCode) && !langjson.HasLanguage(langjson.EmbeddedPack, langCode) {
		return languageJsonNotFound("language_code"), nil
	}
	data, ok := this.registered[query.ModuleName][langCode]
	if !ok {
		data, ok = langjson.EmbeddedPack[query.ModuleName][langCode]
	}
	if !ok {
		return languageJsonNotFound("module_name"), nil
	}
	if query.ModuleName == langjson.CommonModuleName {
		data = langjson.MergedCommonMessages(data, langCode, langjson.EmbeddedPack, this.registered)
	}
	return &it.GetLanguageJsonResult{Data: data, HasData: true}, nil
}

func languageJsonNotFound(field string) *it.GetLanguageJsonResult {
	cerrs := ft.NewClientErrors()
	cerrs.Append(*ft.NewNotFoundError(field))
	return &it.GetLanguageJsonResult{ClientErrors: *cerrs}
}

// RegisterModuleLanguage runs once per module at startup with no request behind it, so like a
// cross-module port it asserts no permission; no REST route reaches it.
func (this *LanguageApplicationServiceImpl) RegisterModuleLanguage(
	moduleName string, translation model.LanguageTranslation,
) error {
	if moduleName == "" {
		return errors.New("registering a language translation requires a module name")
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// Checked before anything is stored, so a rejected module leaves no partial translation behind.
	for langCode := range translation {
		if !model.IsBCP47LanguageCode(langCode) {
			return errors.Errorf("module '%s' has translation for invalid language code '%s'", moduleName, langCode)
		}
		if this.hasTranslation(moduleName, langCode) {
			return errors.Errorf("translation of module '%s' for '%s' is already registered", moduleName, langCode)
		}
	}
	if this.registered[moduleName] == nil {
		this.registered[moduleName] = make(model.LanguageTranslation)
	}
	for langCode, item := range translation {
		this.registered[moduleName][langCode] = item
	}
	return nil
}

func (this *LanguageApplicationServiceImpl) hasTranslation(moduleName string, langCode model.LanguageCode) bool {
	if _, ok := langjson.EmbeddedPack[moduleName][langCode]; ok {
		return true
	}
	_, ok := this.registered[moduleName][langCode]
	return ok
}

func (this *LanguageApplicationServiceImpl) SearchLanguages(ctx corectx.Context, query it.SearchLanguagesQuery) (*it.SearchLanguagesResult, error) {
	return this.languageSvc.SearchLanguages(ctx, query)
}

func (this *LanguageApplicationServiceImpl) UpdateLanguage(ctx corectx.Context, cmd it.UpdateLanguageCommand) (*it.UpdateLanguageResult, error) {
	return this.languageSvc.UpdateLanguage(ctx, cmd)
}
