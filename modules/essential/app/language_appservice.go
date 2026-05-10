package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

func NewLanguageApplicationServiceImpl(languageSvc it.LanguageDomainService) it.LanguageAppService {
	return &LanguageApplicationServiceImpl{languageSvc: languageSvc}
}

type LanguageApplicationServiceImpl struct {
	languageSvc it.LanguageDomainService
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

func (this *LanguageApplicationServiceImpl) SearchLanguages(ctx corectx.Context, query it.SearchLanguagesQuery) (*it.SearchLanguagesResult, error) {
	return this.languageSvc.SearchLanguages(ctx, query)
}

func (this *LanguageApplicationServiceImpl) UpdateLanguage(ctx corectx.Context, cmd it.UpdateLanguageCommand) (*it.UpdateLanguageResult, error) {
	return this.languageSvc.UpdateLanguage(ctx, cmd)
}
