package organization

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type OrganizationRepository interface {
	dyn.BaseRepoGetter
	Insert(ctx corectx.Context, org domain.Organization) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Organization], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Organization]], error)
	Update(ctx corectx.Context, org domain.Organization) (*dyn.OpResult[dyn.MutateResultData], error)
}
