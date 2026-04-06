package role

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type RoleRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Role) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Role) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row domain.Role) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Role], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Role]], error)
	Update(ctx corectx.Context, row domain.Role) (*dyn.OpResult[dyn.MutateResultData], error)
	HasAssignedUsers(ctx corectx.Context, roleId model.Id) bool
	HasAssignedGroups(ctx corectx.Context, roleId model.Id) bool
}
