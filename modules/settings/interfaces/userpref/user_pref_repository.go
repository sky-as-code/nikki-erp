package userpreference

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	domain "github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
)

type UserPreferenceRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.UserPreference) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row domain.UserPreference) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.UserPreference], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.UserPreference]], error)
	Update(ctx corectx.Context, row domain.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error)
}
