package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type RoleRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Role) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Role) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row models.Role) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Role], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Role]], error)
	Update(ctx corectx.Context, row models.Role) (*dyn.OpResult[dyn.MutateResultData], error)
	HasAssignedUsers(ctx corectx.Context, roleId model.Id) bool
	HasAssignedGroups(ctx corectx.Context, roleId model.Id) bool
}
