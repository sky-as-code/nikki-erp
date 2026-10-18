package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type userEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_user"`
}

// registerUserEngine declares the iam_user onion and publishes its typed layers. User is the first
// IAM resource on the composable engine; the others still go through engineSpecs.
func registerUserEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.UserSchemaName),
		func(param composable.BuildParam, storage filestorage.FileStorageAdapter) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.UserSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewUserCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewUserCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewUserCrudApplicationService(base, storage)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p userEngineParam) itUser.UserCrudRepository {
			return p.Engine.Repository().(itUser.UserCrudRepository)
		},
		func(p userEngineParam) itUser.UserCrudDomainService {
			return p.Engine.DomainService().(itUser.UserCrudDomainService)
		},
		func(p userEngineParam) itUser.UserCrudApplicationService {
			return p.Engine.ApplicationService().(itUser.UserCrudApplicationService)
		},
	))
}
