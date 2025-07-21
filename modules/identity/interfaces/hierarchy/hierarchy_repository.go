package hierarchy

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type HierarchyRepository interface {
	AddRemoveUsers(ctx context.Context, param AddRemoveUsersParam) (*ft.ClientError, error)
	Create(ctx context.Context, hierarchyLevel domain.HierarchyLevel) (*domain.HierarchyLevel, error)
	DeleteHard(ctx context.Context, id model.Id) (int, error)
	FindById(ctx context.Context, param FindByIdParam) (*domain.HierarchyLevel, error)
	FindByName(ctx context.Context, param FindByNameParam) (*domain.HierarchyLevel, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.HierarchyLevel], error)
	Update(ctx context.Context, hierarchyLevel domain.HierarchyLevel, prevEtag model.Etag) (*domain.HierarchyLevel, error)
}

type AddRemoveUsersParam = AddRemoveUsersCommand
type DeleteParam = DeleteHierarchyLevelCommand
type FindByIdParam = GetHierarchyLevelByIdQuery
type FindByNameParam struct {
	Name string
}
type SearchParam struct {
	Predicate      *orm.Predicate
	Order          []orm.OrderOption
	Page           int
	Size           int
	IncludeDeleted bool
	WithOrg        bool
	WithChildren   bool
	WithParent     bool
}
