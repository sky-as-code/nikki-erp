package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	userpref "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

type UserPreferenceUiDomainService interface {
	GetUiSavedSearch(ctx corectx.Context, query GetUiSavedSearchQuery) (*GetUiSavedSearchResult, error)
}

type GetUiSavedSearchQuery = userpref.GetUiSavedSearchQuery
type GetUiSavedSearchResult = userpref.GetUiSavedSearchResult
