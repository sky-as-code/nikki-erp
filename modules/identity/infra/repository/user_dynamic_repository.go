package repository

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
	"go.uber.org/dig"
)

type UserDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewUserDynamicRepository(param UserDynamicRepositoryParam) it.UserRepository2 {
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
	dynamicRepo coredyn.BaseRepository
}

func (this *UserDynamicRepository) GetBaseRepo() coredyn.BaseRepository {
	return this.dynamicRepo
}

func (this *UserDynamicRepository) Create(ctx corectx.Context, user domain.UserEntity) (
	*crud.OpResult[domain.UserEntity], error,
) {
	return baserepo.Insert[domain.UserEntity](ctx, this.dynamicRepo, user)
}

func (this *UserDynamicRepository) Update(ctx corectx.Context, user domain.UserEntity, prevEtag string) (
	*crud.OpResult[domain.UserEntity], error,
) {
	return baserepo.Update[domain.UserEntity](ctx, this.dynamicRepo, user, prevEtag)
}

func (this *UserDynamicRepository) FindOne(ctx corectx.Context, param coredyn.GetOneParam) (
	*crud.OpResult[domain.UserEntity], error,
) {
	return baserepo.FindOne[domain.UserEntity](ctx, this.dynamicRepo, param)
}

func (this *UserDynamicRepository) Archive(ctx corectx.Context, user domain.UserEntity) (
	*crud.OpResult[domain.UserEntity], error,
) {
	return baserepo.Archive[domain.UserEntity](ctx, this.dynamicRepo, user)
}

func (this *UserDynamicRepository) Search(
	ctx corectx.Context, param coredyn.SearchParam,
) (*crud.OpResult[crud.PagedResult[domain.UserEntity]], error) {
	return baserepo.Search[domain.UserEntity](ctx, this.dynamicRepo, param)
}
