package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type OrgUnitRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.OrganizationalUnit) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.OrganizationalUnit) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, level domain.OrganizationalUnit) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.OrganizationalUnit], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.OrganizationalUnit]], error)
	Update(ctx corectx.Context, level domain.OrganizationalUnit) (*dyn.OpResult[dyn.MutateResultData], error)
}
