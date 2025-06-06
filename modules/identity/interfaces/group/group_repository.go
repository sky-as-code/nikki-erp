package group

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type GroupRepository interface {
	Create(ctx context.Context, group domain.Group) (*domain.GroupWithOrg, error)
	Update(ctx context.Context, group domain.Group) (*domain.GroupWithOrg, error)
	Delete(ctx context.Context, id model.Id) error
	FindById(ctx context.Context, id model.Id, withOrg bool) (*domain.GroupWithOrg, error)
	FindByName(ctx context.Context, name string) (*domain.GroupWithOrg, error)
	Search(ctx context.Context, criteria *orm.SearchGraph, opts *crud.PagingOptions) (*crud.PagedResult[*domain.Group], error)
}
