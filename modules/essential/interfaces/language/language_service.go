package language

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type LanguageDomainService interface {
	CreateLanguage(ctx corectx.Context, cmd CreateLanguageCommand) (*CreateLanguageResult, error)
	DeleteLanguage(ctx corectx.Context, cmd DeleteLanguageCommand) (*DeleteLanguageResult, error)
	LanguageExists(ctx corectx.Context, query LanguageExistsQuery) (*LanguageExistsResult, error)
	GetLanguage(ctx corectx.Context, query GetLanguageQuery) (*GetLanguageResult, error)
	SearchLanguages(ctx corectx.Context, query SearchLanguagesQuery) (*SearchLanguagesResult, error)
	UpdateLanguage(ctx corectx.Context, cmd UpdateLanguageCommand) (*UpdateLanguageResult, error)
}

type LanguageAppService interface {
	CreateLanguage(ctx corectx.Context, cmd CreateLanguageCommand) (*CreateLanguageResult, error)
	DeleteLanguage(ctx corectx.Context, cmd DeleteLanguageCommand) (*DeleteLanguageResult, error)
	LanguageExists(ctx corectx.Context, query LanguageExistsQuery) (*LanguageExistsResult, error)
	GetLanguage(ctx corectx.Context, query GetLanguageQuery) (*GetLanguageResult, error)
	SearchLanguages(ctx corectx.Context, query SearchLanguagesQuery) (*SearchLanguagesResult, error)
	UpdateLanguage(ctx corectx.Context, cmd UpdateLanguageCommand) (*UpdateLanguageResult, error)
	// GetCurrentLangCode(ctx context.Context) (*GetCurrentLangCodeResult, error)
	// GetCurrentLanguage(ctx context.Context) (*GetCurrentLanguageResult, error)
	// ListEnabledLangCodes(ctx context.Context, query ListEnabledLangCodesQuery) (*ListEnabledLangCodesResult, error)
	// ListLanguages(ctx context.Context, query ListLanguagesQuery) (*ListLanguagesResult, error)
}
