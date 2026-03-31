package group

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type GroupRepository interface {
	dyn.BaseRepoGetter
	Insert(ctx corectx.Context, group domain.Group) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Group], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Group]], error)
	Update(ctx corectx.Context, group domain.Group) (*dyn.OpResult[dyn.MutateResultData], error)
}
