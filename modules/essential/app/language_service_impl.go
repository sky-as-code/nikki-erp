package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

func NewLanguageServiceImpl(repo it.LanguageRepository) it.LanguageService {
	return &LanguageServiceImpl{repo: repo}
}

type LanguageServiceImpl struct{ repo it.LanguageRepository }

func (this *LanguageServiceImpl) CreateLanguage(ctx corectx.Context, cmd it.CreateLanguageCommand) (*it.CreateLanguageResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.Language, *domain.Language]{
		Action: "create language", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *LanguageServiceImpl) DeleteLanguage(ctx corectx.Context, cmd it.DeleteLanguageCommand) (*it.DeleteLanguageResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete language", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *LanguageServiceImpl) LanguageExists(ctx corectx.Context, query it.LanguageExistsQuery) (*it.LanguageExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if language exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *LanguageServiceImpl) GetLanguage(ctx corectx.Context, query it.GetLanguageQuery) (*it.GetLanguageResult, error) {
	return corecrud.GetOne[domain.Language](ctx, corecrud.GetOneParam{
		Action: "get language", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

func (this *LanguageServiceImpl) SearchLanguages(ctx corectx.Context, query it.SearchLanguagesQuery) (*it.SearchLanguagesResult, error) {
	return corecrud.Search[domain.Language](ctx, corecrud.SearchParam{
		Action: "search languages", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *LanguageServiceImpl) UpdateLanguage(ctx corectx.Context, cmd it.UpdateLanguageCommand) (*it.UpdateLanguageResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.Language, *domain.Language]{
		Action: "update language", DbRepoGetter: this.repo, Data: cmd,
	})
}
