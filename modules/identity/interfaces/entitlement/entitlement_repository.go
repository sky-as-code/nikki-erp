package entitlement

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type EntitlementRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.Entitlement) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.Entitlement) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, row models.Entitlement) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.Entitlement], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.Entitlement]], error)
	Update(ctx corectx.Context, row models.Entitlement) (*dyn.OpResult[dyn.MutateResultData], error)
}
