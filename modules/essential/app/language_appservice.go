package app

import (
	"sync"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	langjson "github.com/sky-as-code/nikki-erp/modules/essential/infra/langJson"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewLanguageApplicationService is handed the composable default by the language onion. The
// domain service behind it is the module's own, so the assertion is a wiring check.
func NewLanguageApplicationService(base dynamicresource.CrudApplicationService) it.LanguageApplicationService {
	languageSvc, ok := base.DomainService().(it.LanguageDomainService)
	if !ok {
		panic(errors.New("the language onion must be built with NewLanguageDomainService"))
	}
	return newLanguageApplicationService(base, languageSvc)
}

func newLanguageApplicationService(
	base dynamicresource.CrudApplicationService, languageSvc it.LanguageDomainService,
) *LanguageApplicationServiceImpl {
	return &LanguageApplicationServiceImpl{
		CrudApplicationService: base,
		languageSvc:            languageSvc,
		registered:             make(model.LanguagePack),
	}
}

// LanguageApplicationServiceImpl is the authorized CRUD of the language catalogue (tenant scope,
// the seeded create/update/delete actions), plus the cross-module capability and the translation
// bundles modules register at start-up.
//
// The lookups and the two public actions assert no permission: the lookups are another module
// acting on its own authority, and the enabled list and the bundles render the sign-in page.
type LanguageApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	languageSvc it.LanguageDomainService
	// registered holds the translations modules ship themselves; langjson.EmbeddedPack stays the
	// baseline.
	registered model.LanguagePack
	mutex      sync.RWMutex
}

func (this *LanguageApplicationServiceImpl) GetLanguageByLocaleCode(
	ctx corectx.Context, query it.GetLanguageByLocaleCodeQuery,
) (*it.GetLanguageByLocaleCodeResult, error) {
	return this.languageSvc.GetLanguageByLocaleCode(ctx, query)
}

func (this *LanguageApplicationServiceImpl) ListEnabledLanguages(
	ctx corectx.Context, query it.ListEnabledLanguagesQuery,
) (*it.ListEnabledLanguagesResult, error) {
	return this.languageSvc.ListEnabledLanguages(ctx, query)
}

func (this *LanguageApplicationServiceImpl) ListEnabledLanguagesAction(
	ctx corectx.Context, _ dmodel.DynamicFields,
) (*it.ListEnabledLanguagesResult, error) {
	return this.ListEnabledLanguages(ctx, it.ListEnabledLanguagesQuery{})
}

func (this *LanguageApplicationServiceImpl) GetLanguageJsonAction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*it.GetLanguageJsonResult, error) {
	languageCode, _ := params["language_code"].(string)
	moduleName, _ := params["module_name"].(string)
	return this.GetLanguageJson(ctx, it.GetLanguageJsonQuery{LanguageCode: languageCode, ModuleName: moduleName})
}

func (this *LanguageApplicationServiceImpl) GetLanguageJson(
	ctx corectx.Context, query it.GetLanguageJsonQuery,
) (*it.GetLanguageJsonResult, error) {
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

// Search answers per organization like every other language action when the query names an
// org_id. Without one it is the only org-less language read: it may still only see what a reader
// with no organization may see, the languages marked public_use.
func (this *LanguageApplicationServiceImpl) Search(
	ctx corectx.Context, query dynamicresource.SearchQuery,
) (*dynamicresource.SearchResult, error) {
	if orgId, _ := query[basemodel.FieldOrgId].(string); orgId != "" {
		return this.CrudApplicationService.Search(ctx, query)
	}
	if cErrs := this.AssertPermission(ctx, dynamicresource.PermissionRead, nil); cErrs != nil {
		return &dynamicresource.SearchResult{ClientErrors: *cErrs}, nil
	}
	publicOnly := dmodel.NewSearchNode()
	publicOnly.NewCondition(models.LanguageFieldPublicUse, dmodel.Equals, true)
	graph, err := dmodel.AndDecodedGraph(query["graph"], *publicOnly)
	if err != nil {
		return nil, errors.Wrap(err, "search public languages")
	}
	query["graph"] = graph
	return this.DomainService().Search(ctx, query)
}
