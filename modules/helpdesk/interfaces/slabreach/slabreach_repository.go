package slabreach

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

type SlaBreachRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.SlaBreach) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.SlaBreach) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data models.SlaBreach) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.SlaBreach], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.SlaBreach]], error)
	Update(ctx corectx.Context, data models.SlaBreach) (*dyn.OpResult[dyn.MutateResultData], error)
}
