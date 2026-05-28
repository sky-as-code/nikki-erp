package user

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

// type UserRepository interface {
// 	Create(ctx corecrud.Context, user *domain.User) (*domain.User, error)
// 	// DeleteHard(ctx corecrud.Context, param DeleteParam) (int, error)
// 	Exists(ctx corecrud.Context, id model.Id) (bool, error)
// 	ExistsMulti(ctx corecrud.Context, ids []model.Id, orgId *model.Id) (existing []model.Id, notExisting []model.Id, err error)
// 	FindById(ctx corecrud.Context, param FindByIdParam) (*domain.User, error)
// 	FindByIdForUpdate(ctx corecrud.Context, param FindByIdParam) (*domain.User, *db.DbLock, error)
// 	FindByEmail(ctx corecrud.Context, param FindByEmailParam) (*domain.User, error)
// 	// FindByOrgUnitId(ctx crud.Context, param FindByOrgUnitIdParam) ([]domain.User, error)
// 	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
// 	Search(ctx corecrud.Context, param SearchParam) (*corecrud.PagedResult[domain.User], error)
// 	Update(ctx corecrud.Context, user *domain.User, prevEtag model.Etag) (*domain.User, error)
// }

type UserRepository interface {
	dyn.DynamicModelRepository
	DeleteOne(ctx corectx.Context, keys models.User) (*dyn.OpResult[dyn.MutateResultData], error)
	Exists(ctx corectx.Context, keys []models.User) (*dyn.OpResult[dyn.RepoExistsResult], error)
	Insert(ctx corectx.Context, user models.User) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[models.User], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.User]], error)
	Update(ctx corectx.Context, user models.User) (*dyn.OpResult[dyn.MutateResultData], error)
}

// type DeleteParam = DeleteUserCommand
// type ExistsParam = UserExistsQuery
// type ExistsMultiParam = UserExistsMultiQuery
// type FindByIdParam = GetUserQuery
// type FindByEmailParam = GetUserByEmailQuery
// type FindByOrgUnitIdParam struct {
// 	OrgUnitId model.Id
// 	Status      *domain.UserStatus
// }
// type SearchParam struct {
// 	Predicate     *orm.Predicate
// 	Order         []orm.OrderOption
// 	Page          int
// 	Size          int
// 	WithGroups    bool
// 	WithOrgUnit bool
// 	WithOrgs      bool
// 	OrgId         *model.Id
// }
