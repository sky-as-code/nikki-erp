package user

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type UserRepository interface {
	Create(ctx corecrud.Context, user *domain.User) (*domain.User, error)
	// DeleteHard(ctx corecrud.Context, param DeleteParam) (int, error)
	Exists(ctx corecrud.Context, id model.Id) (bool, error)
	ExistsMulti(ctx corecrud.Context, ids []model.Id, orgId *model.Id) (existing []model.Id, notExisting []model.Id, err error)
	FindById(ctx corecrud.Context, param FindByIdParam) (*domain.User, error)
	FindByIdForUpdate(ctx corecrud.Context, param FindByIdParam) (*domain.User, *db.DbLock, error)
	FindByEmail(ctx corecrud.Context, param FindByEmailParam) (*domain.User, error)
	// FindByHierarchyId(ctx crud.Context, param FindByHierarchyIdParam) ([]domain.User, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx corecrud.Context, param SearchParam) (*corecrud.PagedResult[domain.User], error)
	Update(ctx corecrud.Context, user *domain.User, prevEtag model.Etag) (*domain.User, error)
}

type UserRepository2 interface {
	dEnt.BaseRepoGetter
	Create(ctx corectx.Context, user domain.UserEntity) (*crud.OpResult[domain.UserEntity], error)
	Update(ctx corectx.Context, user domain.UserEntity) (*crud.OpResult[domain.UserEntity], error)
	GetOne(ctx corectx.Context, param dEnt.RepoGetOneParam) (*crud.OpResult[domain.UserEntity], error)
	Search(ctx corectx.Context, param dEnt.RepoSearchParam) (*crud.OpResult[crud.PagedResultData[domain.UserEntity]], error)
}

type DeleteParam = DeleteUserCommand
type ExistsParam = UserExistsQuery
type ExistsMultiParam = UserExistsMultiQuery
type FindByIdParam = GetUserQuery
type FindByEmailParam = GetUserByEmailQuery
type FindByHierarchyIdParam struct {
	HierarchyId model.Id
	Status      *domain.UserStatus
}
type SearchParam struct {
	Predicate     *orm.Predicate
	Order         []orm.OrderOption
	Page          int
	Size          int
	WithGroups    bool
	WithHierarchy bool
	WithOrgs      bool
	OrgId         *model.Id
}
