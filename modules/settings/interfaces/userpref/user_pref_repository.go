package userpreference

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
)

type UserPreferenceRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.UserPreference) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row models.UserPreference) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.UserPreference], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.UserPreference]], error)
	Update(ctx corectx.Context, row models.UserPreference) (*dyn.OpResult[dyn.MutateResultData], error)
}
