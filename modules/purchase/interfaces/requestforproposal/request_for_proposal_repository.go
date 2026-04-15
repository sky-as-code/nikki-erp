package requestforproposal

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

type RequestForProposalRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.RequestForProposal) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, input domain.RequestForProposal) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.RequestForProposal], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.RequestForProposal]], error)
	Update(ctx corectx.Context, input domain.RequestForProposal) (*dyn.OpResult[dyn.MutateResultData], error)
}
