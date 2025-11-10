package module

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

type ModuleRepository interface {
	AcquireLock(ctx crud.Context) (bool, error)
	ReleaseLock(ctx crud.Context) error
	IncludeTransaction(ctx crud.Context) (crud.Context, error)

	Create(ctx crud.Context, module *domain.ModuleMetadata) (*domain.ModuleMetadata, error)
	CreateBulk(ctx crud.Context, modules []*domain.ModuleMetadata) ([]*domain.ModuleMetadata, error)
	DeleteById(ctx crud.Context, param DeleteByIdParam) (int, error)
	Exists(ctx crud.Context, param ExistsParam) (bool, error)
	ExistsByName(ctx crud.Context, param ExistsByNameParam) (bool, error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.ModuleMetadata, error)
	FindByName(ctx crud.Context, param FindByNameParam) (*domain.ModuleMetadata, error)
	List(ctx crud.Context, param ListParam) ([]domain.ModuleMetadata, error)
	Update(ctx crud.Context, module *domain.ModuleMetadata, prevEtag model.Etag) (*domain.ModuleMetadata, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.ModuleMetadata], error)
}

type DeleteByIdParam = DeleteModuleCommand
type ExistsParam = ModuleExistsQuery
type ExistsByNameParam = ModuleExistsByNameQuery
type FindByIdParam = GetModuleByIdQuery
type FindByNameParam = GetModuleByNameQuery
type ListParam = ListModulesQuery
type SearchParam struct {
	Predicate  *orm.Predicate
	Order      []orm.OrderOption
	Page       int
	Size       int
	TypePrefix *string
}
