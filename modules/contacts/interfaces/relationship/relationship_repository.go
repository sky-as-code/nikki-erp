package relationship

import (
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type RelationshipRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys domain.Relationship) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []domain.Relationship) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, relationship domain.Relationship) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.Relationship], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[domain.Relationship]], error)
	Update(ctx corectx.Context, relationship domain.Relationship) (*dyn.OpResult[dyn.MutateResultData], error)
}
