package app

import (
	"strconv"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	tz "github.com/sky-as-code/nikki-erp/modules/essential/domain/timezone"
	langjson "github.com/sky-as-code/nikki-erp/modules/essential/infra/langJson"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/timezone"
)

// timezoneNamespace is the translation bundle holding one label per zone id.
const timezoneNamespace = "timezone"

func NewTimezoneAppServiceImpl() it.TimezoneAppService {
	return &TimezoneAppServiceImpl{}
}

type TimezoneAppServiceImpl struct{}

func (this *TimezoneAppServiceImpl) SearchTimezones(
	_ corectx.Context, query it.SearchTimezonesQuery,
) (*it.SearchTimezonesResult, error) {
	labels, ok := langjson.EmbeddedPack[timezoneNamespace][model.LanguageCode(query.LanguageCode)]
	if !ok {
		labels = langjson.EmbeddedPack[timezoneNamespace][model.LanguageCodeEnUs]
	}
	items, total := tz.Search(query.Keyword, labels, query.Page, query.Size)
	return &it.SearchTimezonesResult{
		Data:    it.SearchTimezonesResultData{Items: items, Total: total},
		HasData: true,
	}, nil
}

func (this *TimezoneAppServiceImpl) SearchTimezonesAction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*it.SearchTimezonesResult, error) {
	keyword, _ := params["keyword"].(string)
	languageCode, _ := params["language_code"].(string)
	return this.SearchTimezones(ctx, it.SearchTimezonesQuery{
		Keyword: keyword, LanguageCode: languageCode, Page: intParam(params["page"]), Size: intParam(params["size"]),
	})
}

// intParam reads a query-string integer, treating an absent or malformed one as zero.
func intParam(value any) int {
	text, _ := value.(string)
	number, _ := strconv.Atoi(text)
	return number
}
