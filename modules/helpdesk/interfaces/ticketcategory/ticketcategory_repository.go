package ticketcategory

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

type TicketCategoryRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.TicketCategory) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.TicketCategory) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data domain.TicketCategory) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.TicketCategory], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.TicketCategory]], error)
	Update(ctx corectx.Context, data domain.TicketCategory) (*dyn.OpResult[dyn.MutateResultData], error)
}
