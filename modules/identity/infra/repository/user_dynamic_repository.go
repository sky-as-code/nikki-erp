package repository

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	schemaEnt "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity/baserepo"
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
			Schema:       schemaEnt.MustGetSchema(domain.UserSchemaName),
		},
	)
	return &UserDynamicRepository{dynamicRepo: dynamicRepo}
}

type UserDynamicRepository struct {
	dynamicRepo dEnt.BaseRepository
}

func (this *UserDynamicRepository) GetBaseRepo() dynamicentity.BaseRepository {
	return this.dynamicRepo
}

func (this *UserDynamicRepository) Create(ctx dEnt.Context, user domain.UserEntity) (
	*dEnt.OpResult[domain.UserEntity], error,
) {
	return baserepo.Insert[domain.UserEntity](ctx, this.dynamicRepo, user)
}

func (this *UserDynamicRepository) Update(ctx dEnt.Context, user domain.UserEntity, prevEtag string) (
	*dEnt.OpResult[domain.UserEntity], error,
) {
	return baserepo.Update[domain.UserEntity](ctx, this.dynamicRepo, user, prevEtag)
}

func (this *UserDynamicRepository) FindOne(ctx dEnt.Context, param dEnt.GetOneParam) (
	*dEnt.OpResult[domain.UserEntity], error,
) {
	return baserepo.FindOne[domain.UserEntity](ctx, this.dynamicRepo, param)
}

func (this *UserDynamicRepository) Archive(ctx dEnt.Context, user domain.UserEntity) (
	*dEnt.OpResult[domain.UserEntity], error,
) {
	return baserepo.Archive[domain.UserEntity](ctx, this.dynamicRepo, user)
}

func (this *UserDynamicRepository) Search(
	ctx dEnt.Context, param dEnt.SearchParam,
) (*dEnt.OpResult[dEnt.PagedResult[domain.UserEntity]], error) {
	return baserepo.Search[domain.UserEntity](ctx, this.dynamicRepo, param)
}
