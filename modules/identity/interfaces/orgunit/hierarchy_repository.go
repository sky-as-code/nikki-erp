package orgunit

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type OrgUnitRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.OrganizationalUnit) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.OrganizationalUnit) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, level models.OrganizationalUnit) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.OrganizationalUnit], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.OrganizationalUnit]], error)
	Update(ctx corectx.Context, level models.OrganizationalUnit) (*dyn.OpResult[dyn.MutateResultData], error)
}
