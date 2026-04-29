package userpreference

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type UserPreferenceCrudDomainService interface {
	CreateUserPreference(ctx corectx.Context, cmd CreateUserPreferenceCommand) (*CreateUserPreferenceResult, error)
	DeleteUserPreference(ctx corectx.Context, cmd DeleteUserPreferenceCommand) (*DeleteUserPreferenceResult, error)
	GetUserPreference(ctx corectx.Context, query GetUserPreferenceQuery) (*GetUserPreferenceResult, error)
	SearchUserPreferences(ctx corectx.Context, query SearchUserPreferencesQuery) (*SearchUserPreferencesResult, error)
	UserPreferenceExists(ctx corectx.Context, query UserPreferenceExistsQuery) (*UserPreferenceExistsResult, error)
	UpdateUserPreference(ctx corectx.Context, cmd UpdateUserPreferenceCommand) (*UpdateUserPreferenceResult, error)
}

type UserPreferenceUiDomainService interface {
	GetUiSavedSearch(ctx corectx.Context, query GetUiSavedSearchQuery) (*GetUiSavedSearchResult, error)
}

type UserPreferenceApplicationService interface {
	CreateUserPreference(ctx corectx.Context, cmd CreateUserPreferenceCommand) (*CreateUserPreferenceResult, error)
	DeleteUserPreference(ctx corectx.Context, cmd DeleteUserPreferenceCommand) (*DeleteUserPreferenceResult, error)
	UserPreferenceExists(ctx corectx.Context, query UserPreferenceExistsQuery) (*UserPreferenceExistsResult, error)
	GetUserPreference(ctx corectx.Context, query GetUserPreferenceQuery) (*GetUserPreferenceResult, error)
	SearchUserPreferences(ctx corectx.Context, query SearchUserPreferencesQuery) (*SearchUserPreferencesResult, error)
	UpdateUserPreference(ctx corectx.Context, cmd UpdateUserPreferenceCommand) (*UpdateUserPreferenceResult, error)
}
