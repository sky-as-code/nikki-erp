package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itService "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/service"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type serviceEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_service"`
}

// registerServiceEngine declares the iam_service onion. Its repository is the user repository
// viewed through the iam_service schema: users and services share iam_users, and the domain
// services - not the repository - decide which type each resource sees.
func registerServiceEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ServiceSchemaName),
		func(
			param composable.BuildParam, userRepo itUser.UserCrudRepository,
			permRepo itPerm.PermissionRepository, historyRepo itPerm.PermissionHistoryRepository,
		) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.ServiceSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewServiceCrudRepository(userRepo, base.Schema())
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewServiceCrudDomainService(base, permRepo, historyRepo)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					domSvc := base.DomainService().(itService.ServiceCrudDomainService)
					return app.NewServiceCrudApplicationService(base, domSvc)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p serviceEngineParam) itService.ServiceCrudRepository {
			return p.Engine.Repository().(itService.ServiceCrudRepository)
		},
		func(p serviceEngineParam) itService.ServiceCrudDomainService {
			return p.Engine.DomainService().(itService.ServiceCrudDomainService)
		},
		func(p serviceEngineParam) itService.ServiceCrudApplicationService {
			return p.Engine.ApplicationService().(itService.ServiceCrudApplicationService)
		},
	))
}
