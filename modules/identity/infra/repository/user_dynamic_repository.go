package repository

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicentity/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewUserDynamicRepository(client orm.DbClient, queryBuilder orm.QueryBuilder) it.UserRepository2 {
	dynamicRepo := dynamicentity.NewDbRepositoryImpl(client, queryBuilder, schema.MustGetSchema(domain.UserSchemaName))
	return &UserDynamicRepository{
		dynamicRepo: dynamicRepo,
	}
}

type UserDynamicRepository struct {
	dynamicRepo dEnt.DbRepository
}

// Implements dynamicentity.DbRepoGetter interface
func (this *UserDynamicRepository) GetDbRepo() dynamicentity.DbRepository {
	return this.dynamicRepo
}

// Implements it.UserRepository2 interface
func (this *UserDynamicRepository) Create(ctx dEnt.Context, user domain.UserEntity) (*domain.UserEntity, error) {
	return baserepo.Insert[domain.UserEntity](ctx, this.dynamicRepo, user)
}
