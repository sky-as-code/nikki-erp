package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type userEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_user"`
}

// registerUserEngine declares the iam_user onion and publishes its typed layers. User is the first
// IAM resource on the composable engine; the others still go through engineSpecs.
func registerUserEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.UserSchemaName),
		func(
			param dynamicresource.BuildParam,
			settingsSvc itExt.UserSettingsExtService, permRepo itPerm.PermissionRepository,
			historyRepo itPerm.PermissionHistoryRepository,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.UserSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewUserCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewUserCrudDomainService(base, settingsSvc, permRepo, historyRepo)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					domSvc := base.DomainService().(itUser.UserCrudDomainService)
					return app.NewUserCrudApplicationService(base, domSvc)
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
