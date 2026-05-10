package vendor

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

type VendorRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Vendor) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Vendor) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, input models.Vendor) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Vendor], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Vendor]], error)
	Update(ctx corectx.Context, input models.Vendor) (*dyn.OpResult[dyn.MutateResultData], error)
}
