package repository

import (
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type UserDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewUserDynamicRepository(param UserDynamicRepositoryParam) it.UserRepository {
	dynamicRepo := baserepo.NewBaseRepositoryImpl(
		baserepo.NewBaseRepositoryParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(domain.UserSchemaName),
		},
	)
	return &UserDynamicRepository{dynamicRepo: dynamicRepo}
}

type UserDynamicRepository struct {
	dynamicRepo dyn.BaseRepository
}

func (this *UserDynamicRepository) GetBaseRepo() dyn.BaseRepository {
	return this.dynamicRepo
}

func (this *UserDynamicRepository) Insert(
	ctx corectx.Context, user domain.User,
) (*dyn.OpResult[int], error) {
	return baserepo.Insert(ctx, this.dynamicRepo, user)
}

func (this *UserDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.User], error) {
	return baserepo.GetOne[domain.User](ctx, this.dynamicRepo, param)
}

func (this *UserDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[domain.User]], error) {
	return baserepo.Search[domain.User](ctx, this.dynamicRepo, param)
}

func (this *UserDynamicRepository) Update(ctx corectx.Context, user domain.User) (
	*dyn.OpResult[dyn.MutateResultData], error,
) {
	return baserepo.Update(ctx, this.dynamicRepo, user.GetFieldData())
}
