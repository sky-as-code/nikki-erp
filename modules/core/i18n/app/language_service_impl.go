package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	it "github.com/sky-as-code/nikki-erp/modules/core/i18n/interfaces"
)

func NewLanguageServiceImpl() it.LanguageService {
	return &LanguageServiceImpl{}
}

type LanguageServiceImpl struct {
}

func (this *LanguageServiceImpl) GetCurrentLangCode(ctx context.Context) (*it.GetCurrentLangCodeResult, error) {
	return &it.GetCurrentLangCodeResult{
		Data:    model.DefaultLanguageCode,
		HasData: true,
	}, nil
}

func (this *LanguageServiceImpl) GetCurrentLanguage(ctx context.Context) (*it.GetCurrentLanguageResult, error) {
	return nil, nil
}

func (this *LanguageServiceImpl) ListEnabledLangCodes(ctx context.Context, query it.ListEnabledLangCodesQuery) (*it.ListEnabledLangCodesResult, error) {
	return &it.ListEnabledLangCodesResult{
		Data:    []model.LanguageCode{model.LanguageCode("en-US"), model.LanguageCode("vi-VN")},
		HasData: true,
	}, nil
}

func (this *LanguageServiceImpl) ListLanguages(ctx context.Context, query it.ListLanguagesQuery) (*it.ListLanguagesResult, error) {
	return &it.ListLanguagesResult{
		Data: it.ListLanguagesResultData{
			Items: languages,
			Total: len(languages),
		},
		HasData: true,
	}, nil
}

var languages = []it.Language{
	{
		Id:   util.ToPtr("01JZYF7DT3ASR4PXAX4HP4BVGZ"),
		Name: util.ToPtr("English (US)"),
		Code: util.ToPtr(model.LanguageCode("en-US")),
	},
	{
		Id:   util.ToPtr("01JZYHK1T56TJYSK4A7EJ3A7R9"),
		Name: util.ToPtr("Vietnamese / Tiếng Việt"),
		Code: util.ToPtr(model.LanguageCode("vi-VN")),
	},
}
