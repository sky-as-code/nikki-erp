package commchannel

import (
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type CommChannelRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.CommChannel) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.CommChannel) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, commChannel domain.CommChannel) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.CommChannel], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.CommChannel]], error)
	Update(ctx corectx.Context, commChannel domain.CommChannel) (*dyn.OpResult[dyn.MutateResultData], error)
}
