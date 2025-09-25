package hierarchy

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type HierarchyRepository interface {
	AddRemoveUsers(ctx crud.Context, param AddRemoveUsersParam) (*ft.ClientError, error)
	Create(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel) (*domain.HierarchyLevel, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.HierarchyLevel, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.HierarchyLevel, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.HierarchyLevel], error)
	Update(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel, prevEtag model.Etag) (*domain.HierarchyLevel, error)
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
