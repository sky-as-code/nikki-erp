package ticketmessage

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
)

type TicketMessageRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.TicketMessage) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.TicketMessage) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data models.TicketMessage) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.TicketMessage], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.TicketMessage]], error)
	Update(ctx corectx.Context, data models.TicketMessage) (*dyn.OpResult[dyn.MutateResultData], error)
}
