package ticketassignment

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

type TicketAssignmentRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.TicketAssignment) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.TicketAssignment) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, data domain.TicketAssignment) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.TicketAssignment], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.TicketAssignment]], error)
	Update(ctx corectx.Context, data domain.TicketAssignment) (*dyn.OpResult[dyn.MutateResultData], error)
}
