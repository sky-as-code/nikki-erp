package role_request

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type RoleRequestRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.RoleRequest) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.RoleRequest) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row domain.RoleRequest) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.RoleRequest], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.RoleRequest]], error)
	Update(ctx corectx.Context, row domain.RoleRequest) (*dyn.OpResult[dyn.MutateResultData], error)
}
