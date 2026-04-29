package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

func NewUserPreferenceCrudDomainServiceImpl(
	repo it.UserPreferenceRepository,
) it.UserPreferenceCrudDomainService {
	return &UserPreferenceDomainServiceImpl{repo: repo}
}

type UserPreferenceDomainServiceImpl struct {
	repo it.UserPreferenceRepository
}

func (this *UserPreferenceDomainServiceImpl) CreateUserPreference(
	ctx corectx.Context, cmd it.CreateUserPreferenceCommand,
) (*it.CreateUserPreferenceResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.UserPreference, *domain.UserPreference]{
		Action:         "create user preference",
		BaseRepoGetter: this.repo,
		Data:           cmd,
	})
}

func (this *UserPreferenceDomainServiceImpl) DeleteUserPreference(
	ctx corectx.Context, cmd it.DeleteUserPreferenceCommand,
) (*it.DeleteUserPreferenceResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete user preference",
		DbRepoGetter: this.repo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *UserPreferenceDomainServiceImpl) UserPreferenceExists(
	ctx corectx.Context, query it.UserPreferenceExistsQuery,
) (*it.UserPreferenceExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if user preference exists",
		DbRepoGetter: this.repo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *UserPreferenceDomainServiceImpl) GetUserPreference(
	ctx corectx.Context, query it.GetUserPreferenceQuery,
) (*it.GetUserPreferenceResult, error) {
	return corecrud.GetOne[domain.UserPreference](ctx, corecrud.GetOneParam{
		Action:       "get user preference",
		DbRepoGetter: this.repo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *UserPreferenceDomainServiceImpl) SearchUserPreferences(
	ctx corectx.Context, query it.SearchUserPreferencesQuery,
) (*it.SearchUserPreferencesResult, error) {
	return corecrud.Search[domain.UserPreference](ctx, corecrud.SearchParam{
		Action:       "search user preferences",
		DbRepoGetter: this.repo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *UserPreferenceDomainServiceImpl) UpdateUserPreference(
	ctx corectx.Context, cmd it.UpdateUserPreferenceCommand,
) (*it.UpdateUserPreferenceResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.UserPreference, *domain.UserPreference]{
		Action:       "update user preference",
		DbRepoGetter: this.repo,
		Data:         cmd,
	})
}
