package language

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type LanguageDomainService interface {
	CreateLanguage(ctx corectx.Context, cmd CreateLanguageCommand) (*CreateLanguageResult, error)
	DeleteLanguage(ctx corectx.Context, cmd DeleteLanguageCommand) (*DeleteLanguageResult, error)
	LanguageExists(ctx corectx.Context, query LanguageExistsQuery) (*LanguageExistsResult, error)
	GetLanguage(ctx corectx.Context, query GetLanguageQuery) (*GetLanguageResult, error)
	GetLanguageByIsoCode(ctx corectx.Context, query GetLanguageByIsoCodeQuery) (*GetLanguageByIsoCodeResult, error)
	ListEnabledLanguages(ctx corectx.Context, query ListEnabledLanguagesQuery) (*ListEnabledLanguagesResult, error)
	SearchLanguages(ctx corectx.Context, query SearchLanguagesQuery) (*SearchLanguagesResult, error)
	UpdateLanguage(ctx corectx.Context, cmd UpdateLanguageCommand) (*UpdateLanguageResult, error)
}

type LanguageAppService interface {
	CreateLanguage(ctx corectx.Context, cmd CreateLanguageCommand) (*CreateLanguageResult, error)
	DeleteLanguage(ctx corectx.Context, cmd DeleteLanguageCommand) (*DeleteLanguageResult, error)
	LanguageExists(ctx corectx.Context, query LanguageExistsQuery) (*LanguageExistsResult, error)
	GetLanguage(ctx corectx.Context, query GetLanguageQuery) (*GetLanguageResult, error)
	GetLanguageByIsoCode(ctx corectx.Context, query GetLanguageByIsoCodeQuery) (*GetLanguageByIsoCodeResult, error)
	ListEnabledLanguages(ctx corectx.Context, query ListEnabledLanguagesQuery) (*ListEnabledLanguagesResult, error)
	GetLanguageJson(ctx corectx.Context, query GetLanguageJsonQuery) (*GetLanguageJsonResult, error)
	// RegisterModuleLanguage stores a module's own translation, served by GetLanguageJson under
	// moduleName. A language already embedded or registered for that module is an error.
	RegisterModuleLanguage(moduleName string, translation model.LanguageTranslation) error
	SearchLanguages(ctx corectx.Context, query SearchLanguagesQuery) (*SearchLanguagesResult, error)
	UpdateLanguage(ctx corectx.Context, cmd UpdateLanguageCommand) (*UpdateLanguageResult, error)
	// GetCurrentLangCode(ctx context.Context) (*GetCurrentLangCodeResult, error)
	// GetCurrentLanguage(ctx context.Context) (*GetCurrentLanguageResult, error)
	// ListLanguages(ctx context.Context, query ListLanguagesQuery) (*ListLanguagesResult, error)
}
