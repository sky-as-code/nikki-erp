package slabreach

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

type SlaBreachRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.SlaBreach) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.SlaBreach) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data domain.SlaBreach) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.SlaBreach], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.SlaBreach]], error)
	Update(ctx corectx.Context, data domain.SlaBreach) (*dyn.OpResult[dyn.MutateResultData], error)
}
