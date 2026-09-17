package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

func NewLanguageDomainServiceImpl(repo it.LanguageRepository) it.LanguageDomainService {
	return &LanguageDomainServiceImpl{repo: repo}
}

type LanguageDomainServiceImpl struct{ repo it.LanguageRepository }

func (this *LanguageDomainServiceImpl) CreateLanguage(ctx corectx.Context, cmd it.CreateLanguageCommand) (*it.CreateLanguageResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Language, *models.Language]{
		Action: "create language", BaseRepoGetter: this.repo, Data: cmd,
	})
}

func (this *LanguageDomainServiceImpl) DeleteLanguage(ctx corectx.Context, cmd it.DeleteLanguageCommand) (*it.DeleteLanguageResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action: "delete language", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd),
	})
}

func (this *LanguageDomainServiceImpl) LanguageExists(ctx corectx.Context, query it.LanguageExistsQuery) (*it.LanguageExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action: "check if language exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query),
	})
}

func (this *LanguageDomainServiceImpl) GetLanguage(ctx corectx.Context, query it.GetLanguageQuery) (*it.GetLanguageResult, error) {
	return corecrud.GetOne[models.Language](ctx, corecrud.GetOneParam{
		Action: "get language", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query),
	})
}

// GetLanguageByIsoCode resolves the code a setting or an i18n layer names a language with. The
// iso_code column is unique, so a match is exactly one row.
func (this *LanguageDomainServiceImpl) GetLanguageByIsoCode(
	ctx corectx.Context, query it.GetLanguageByIsoCodeQuery,
) (*it.GetLanguageByIsoCodeResult, error) {
	found, err := this.repo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.LanguageFieldIsoCode: query.IsoCode},
	})
	if err != nil {
		return nil, errors.Wrap(err, "get language by iso code")
	}
	if !found.HasData {
		// Absence is a legitimate answer: the caller decides what an unknown language means.
		return &it.GetLanguageByIsoCodeResult{HasData: false}, nil
	}
	return &it.GetLanguageByIsoCodeResult{Data: found.Data, HasData: true}, nil
}

func (this *LanguageDomainServiceImpl) SearchLanguages(ctx corectx.Context, query it.SearchLanguagesQuery) (*it.SearchLanguagesResult, error) {
	return corecrud.Search[models.Language](ctx, corecrud.SearchParam{
		Action: "search languages", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query),
	})
}

func (this *LanguageDomainServiceImpl) UpdateLanguage(ctx corectx.Context, cmd it.UpdateLanguageCommand) (*it.UpdateLanguageResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Language, *models.Language]{
		Action: "update language", DbRepoGetter: this.repo, Data: cmd,
	})
}
