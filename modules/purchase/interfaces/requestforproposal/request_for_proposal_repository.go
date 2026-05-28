package requestforproposal

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

type RequestForProposalRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.RequestForProposal) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, input models.RequestForProposal) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.RequestForProposal], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.RequestForProposal]], error)
	Update(ctx corectx.Context, input models.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error)
}
