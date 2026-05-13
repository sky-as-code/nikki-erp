package interfaces

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

var getCurrentLangCodeQuery = cqrs.RequestType{
	Module:    "core",
	Submodule: "i18n",
	Action:    "getCurrentLangCode",
}

type GetCurrentLangCodeQuery struct{}

func (GetCurrentLangCodeQuery) CqrsRequestType() cqrs.RequestType {
	return getCurrentLangCodeQuery
}

type GetCurrentLangCodeResult = crud.OpResult[model.LanguageCode]

var getCurrentLanguageQuery = cqrs.RequestType{
	Module:    "core",
	Submodule: "i18n",
	Action:    "getCurrentLanguage",
}

type GetCurrentLanguageQuery struct{}

func (GetCurrentLanguageQuery) CqrsRequestType() cqrs.RequestType {
	return getCurrentLanguageQuery
}

type GetCurrentLanguageResult = crud.OpResult[Language]

var listEnabledLangCodesQuery = cqrs.RequestType{
	Module:    "core",
	Submodule: "i18n",
	Action:    "listEnabledLangCodes",
}

type ListEnabledLangCodesQuery struct{}

func (ListEnabledLangCodesQuery) CqrsRequestType() cqrs.RequestType {
	return listEnabledLangCodesQuery
}

type ListEnabledLangCodesResult = crud.OpResult[[]model.LanguageCode]

var listLanguagesQuery = cqrs.RequestType{
	Module:    "core",
	Submodule: "i18n",
	Action:    "listLanguages",
}

// List all supported languages
type ListLanguagesQuery struct {
	// Whether the language is installed, and enabled or disabled
	IsEnabled *bool `json:"isEnabled" query:"isEnabled"`
	// Whether the language is installed and ready to be enabled
	IsInstalled *bool   `json:"isInstalled" query:"isInstalled"`
	Graph       *string `json:"graph" query:"graph"`
	Page        *int    `json:"page" query:"page"`
	Size        *int    `json:"size" query:"size"`
}

func (ListLanguagesQuery) CqrsRequestType() cqrs.RequestType {
	return listLanguagesQuery
}

type ListLanguagesResultData = crud.PagedResult[Language]
type ListLanguagesResult = crud.OpResult[ListLanguagesResultData]
