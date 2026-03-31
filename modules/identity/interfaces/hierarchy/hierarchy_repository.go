package hierarchy

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type HierarchyRepository interface {
	dyn.BaseRepoGetter
	Insert(ctx corectx.Context, level domain.HierarchyLevel) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.HierarchyLevel], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.HierarchyLevel]], error)
	Update(ctx corectx.Context, level domain.HierarchyLevel) (*dyn.OpResult[dyn.MutateResultData], error)
}
