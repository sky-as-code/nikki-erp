package language

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// LanguageRepository reads and writes language rows: the composable default today.
type LanguageRepository interface {
	composable.CrudRepository
}

// LanguageLookupService is what another module asks of the language catalogue.
type LanguageLookupService interface {
	GetLanguageByIsoCode(ctx corectx.Context, query GetLanguageByIsoCodeQuery) (*GetLanguageByIsoCodeResult, error)
	ListEnabledLanguages(ctx corectx.Context, query ListEnabledLanguagesQuery) (*ListEnabledLanguagesResult, error)
}

// LanguageDomainService is the CRUD of the resource plus the lookups.
type LanguageDomainService interface {
	composable.CrudDomainService
	LanguageLookupService
}

// LanguageAppService is the capability other modules (iam, essential's own settings rule and
// module start-up) bind to. It deliberately excludes the CRUD.
type LanguageAppService interface {
	LanguageLookupService

	GetLanguageJson(ctx corectx.Context, query GetLanguageJsonQuery) (*GetLanguageJsonResult, error)

	// RegisterModuleLanguage stores a module's own translation, served by GetLanguageJson under
	// moduleName. A language already embedded or registered for that module is an error.
	RegisterModuleLanguage(moduleName string, translation model.LanguageTranslation) error
}

// LanguageApplicationService is the whole application layer: the authorized CRUD the REST handler
// serves, the cross-module capability, and the two public collection actions.
type LanguageApplicationService interface {
	composable.CrudApplicationService
	LanguageAppService

	// GetLanguageJsonAction serves GET .../json: public, the translation bundles render every page
	// including the sign-in page.
	GetLanguageJsonAction(ctx corectx.Context, params dmodel.DynamicFields) (*GetLanguageJsonResult, error)

	// ListEnabledLanguagesAction serves GET .../enabled: public, the sign-in page offers a
	// language switch before anyone has a session.
	ListEnabledLanguagesAction(
		ctx corectx.Context, params dmodel.DynamicFields,
	) (*ListEnabledLanguagesResult, error)
}
