package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/userpref"
)

func NewUserPreferenceApplicationServiceImpl(
	domainSvc it.UserPreferenceCrudDomainService,
) it.UserPreferenceApplicationService {
	return &UserPreferenceApplicationServiceImpl{
		domainSvc: domainSvc,
	}
}

type UserPreferenceApplicationServiceImpl struct {
	domainSvc it.UserPreferenceCrudDomainService
}

func (this *UserPreferenceApplicationServiceImpl) CreateUserPreference(
	ctx corectx.Context, cmd it.CreateUserPreferenceCommand,
) (*it.CreateUserPreferenceResult, error) {
	return this.domainSvc.CreateUserPreference(ctx, cmd)
}

func (this *UserPreferenceApplicationServiceImpl) DeleteUserPreference(
	ctx corectx.Context, cmd it.DeleteUserPreferenceCommand,
) (*it.DeleteUserPreferenceResult, error) {
	return this.domainSvc.DeleteUserPreference(ctx, cmd)
}

func (this *UserPreferenceApplicationServiceImpl) UserPreferenceExists(
	ctx corectx.Context, query it.UserPreferenceExistsQuery,
) (*it.UserPreferenceExistsResult, error) {
	return this.domainSvc.UserPreferenceExists(ctx, query)
}

func (this *UserPreferenceApplicationServiceImpl) GetUserPreference(
	ctx corectx.Context, query it.GetUserPreferenceQuery,
) (*it.GetUserPreferenceResult, error) {
	return this.domainSvc.GetUserPreference(ctx, query)
}

func (this *UserPreferenceApplicationServiceImpl) SearchUserPreferences(
	ctx corectx.Context, query it.SearchUserPreferencesQuery,
) (*it.SearchUserPreferencesResult, error) {
	return this.domainSvc.SearchUserPreferences(ctx, query)
}

func (this *UserPreferenceApplicationServiceImpl) UpdateUserPreference(
	ctx corectx.Context, cmd it.UpdateUserPreferenceCommand,
) (*it.UpdateUserPreferenceResult, error) {
	return this.domainSvc.UpdateUserPreference(ctx, cmd)
}
