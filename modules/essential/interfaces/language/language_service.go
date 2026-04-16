package language

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type LanguageService interface {
	CreateLanguage(ctx corectx.Context, cmd CreateLanguageCommand) (*CreateLanguageResult, error)
	DeleteLanguage(ctx corectx.Context, cmd DeleteLanguageCommand) (*DeleteLanguageResult, error)
	LanguageExists(ctx corectx.Context, query LanguageExistsQuery) (*LanguageExistsResult, error)
	GetLanguage(ctx corectx.Context, query GetLanguageQuery) (*GetLanguageResult, error)
	SearchLanguages(ctx corectx.Context, query SearchLanguagesQuery) (*SearchLanguagesResult, error)
	UpdateLanguage(ctx corectx.Context, cmd UpdateLanguageCommand) (*UpdateLanguageResult, error)
}
