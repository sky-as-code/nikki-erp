package services

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

func NewUserPreferenceUiDomainServiceImpl(
	userPrefSvc it.UserPreferenceCrudDomainService,
) it.UserPreferenceUiDomainService {
	return userPrefSvc.(it.UserPreferenceUiDomainService)
}

// Implements UserPreferenceUiDomainService
func (this *UserPreferenceDomainServiceImpl) GetUiSavedSearch(
	ctx corectx.Context, query it.GetUiSavedSearchQuery,
) (*it.GetUiSavedSearchResult, error) {
	return &it.GetUiSavedSearchResult{
		Data: it.GetUiSavedSearchResultData{
			Fields: []string{"id", "display_name", "email", "status"},
		},
	}, nil
}

// Implements nikkierp/modules/core/dynamicmodel/crud/FieldsResolver
func (this *UserPreferenceDomainServiceImpl) GetListFields(
	ctx corectx.Context, uiName string, userId model.Id,
) (*dyn.OpResult[[]string], error) {
	result, _ := this.GetUiSavedSearch(ctx, it.GetUiSavedSearchQuery{
		SearchName: uiName,
		UserId:     userId,
	})
	return &dyn.OpResult[[]string]{
		Data: result.Data.Fields,
	}, nil
}
