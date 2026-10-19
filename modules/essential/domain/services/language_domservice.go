package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewLanguageDomainService is handed the composable default by the language onion. The CRUD is
// the default's; what this type adds is the lookups other modules make.
func NewLanguageDomainService(base composable.CrudDomainService) it.LanguageDomainService {
	return &LanguageDomainServiceImpl{CrudDomainService: base}
}

type LanguageDomainServiceImpl struct {
	composable.CrudDomainService
}

// GetLanguageByIsoCode resolves the code a setting or an i18n layer names a language with. The
// iso_code column is unique, so a match is exactly one row.
func (this *LanguageDomainServiceImpl) GetLanguageByIsoCode(
	ctx corectx.Context, query it.GetLanguageByIsoCodeQuery,
) (*it.GetLanguageByIsoCodeResult, error) {
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.LanguageFieldIsoCode: query.IsoCode},
	})
	if err != nil {
		return nil, errors.Wrap(err, "get language by iso code")
	}
	if !found.HasData {
		// Absence is a legitimate answer: the caller decides what an unknown language means.
		return &it.GetLanguageByIsoCodeResult{HasData: false}, nil
	}
	return &it.GetLanguageByIsoCodeResult{Data: *models.NewLanguageFrom(found.Data), HasData: true}, nil
}

// enabledLanguagesLimit bounds the read; a catalogue of locales the product ships translations for
// stays far below it.
const enabledLanguagesLimit = 500

func (this *LanguageDomainServiceImpl) ListEnabledLanguages(
	ctx corectx.Context, query it.ListEnabledLanguagesQuery,
) (*it.ListEnabledLanguagesResult, error) {
	graph := dmodel.NewSearchGraph().
		NewCondition(models.LanguageFieldIsEnabled, dmodel.Equals, true).
		OrderBy(models.LanguageFieldName)
	found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: enabledLanguagesLimit})
	if err != nil {
		return nil, errors.Wrap(err, "list enabled languages")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "list enabled languages")
	}
	languages := []models.Language{}
	if found.HasData {
		for _, row := range found.Data.Items {
			languages = append(languages, *models.NewLanguageFrom(row))
		}
	}
	return &it.ListEnabledLanguagesResult{Data: languages, HasData: true}, nil
}
