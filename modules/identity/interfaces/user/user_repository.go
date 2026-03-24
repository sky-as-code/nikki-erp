package user

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type UserRepository interface {
	Create(ctx crud.Context, user *domain.User) (*domain.User, error)
	DeleteHard(ctx crud.Context, param DeleteParam) (int, error)
	Exists(ctx crud.Context, id model.Id) (bool, error)
	ExistsMulti(ctx crud.Context, ids []model.Id, orgId *model.Id) (existing []model.Id, notExisting []model.Id, err error)
	FindById(ctx crud.Context, param FindByIdParam) (*domain.User, error)
	FindByIdForUpdate(ctx crud.Context, param FindByIdParam) (*domain.User, *db.DbLock, error)
	FindByEmail(ctx crud.Context, param FindByEmailParam) (*domain.User, error)
	// FindByHierarchyId(ctx crud.Context, param FindByHierarchyIdParam) ([]domain.User, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[domain.User], error)
	Update(ctx crud.Context, user *domain.User, prevEtag model.Etag) (*domain.User, error)
}

type UserRepository2 interface {
	dEnt.BaseRepoGetter
	Create(ctx dEnt.Context, user domain.UserEntity) (*dEnt.OpResult[domain.UserEntity], error)
	Update(ctx dEnt.Context, user domain.UserEntity, prevEtag string) (*dEnt.OpResult[domain.UserEntity], error)
	FindOne(ctx dEnt.Context, param dEnt.GetOneParam) (*dEnt.OpResult[domain.UserEntity], error)
	Search(ctx dEnt.Context, param dEnt.SearchParam) (*dEnt.OpResult[dEnt.PagedResult[domain.UserEntity]], error)
	Archive(ctx dEnt.Context, user domain.UserEntity) (*dEnt.OpResult[domain.UserEntity], error)
}

type DeleteParam = DeleteUserCommand
type ExistsParam = UserExistsQuery
type ExistsMultiParam = UserExistsMultiQuery
type FindByIdParam = GetUser
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
