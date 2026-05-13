package interfaces

import "context"

type LanguageService interface {
	GetCurrentLangCode(ctx context.Context) (*GetCurrentLangCodeResult, error)
	GetCurrentLanguage(ctx context.Context) (*GetCurrentLanguageResult, error)
	ListEnabledLangCodes(ctx context.Context, query ListEnabledLangCodesQuery) (*ListEnabledLangCodesResult, error)
	ListLanguages(ctx context.Context, query ListLanguagesQuery) (*ListLanguagesResult, error)
}
