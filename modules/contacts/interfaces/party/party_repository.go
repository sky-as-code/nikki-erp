package party

import (
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type PartyRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Party) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Party) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, party domain.Party) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Party], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Party]], error)
	Update(ctx corectx.Context, party domain.Party) (*dyn.OpResult[dyn.MutateResultData], error)
}
