package purchaserequest

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

type PurchaseRequestRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.PurchaseRequest) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.PurchaseRequest) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, input models.PurchaseRequest) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.PurchaseRequest], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.PurchaseRequest]], error)
	Update(ctx corectx.Context, input models.PurchaseRequest) (*dyn.OpResult[dyn.MutateResultData], error)
}
