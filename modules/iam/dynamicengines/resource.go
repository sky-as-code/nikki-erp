package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

type resourceEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_resource"`
}

// registerResourceEngine declares the iam_resource onion: the catalog of what can be authorized,
// shared by the whole tenant, so its permissions are tenant-scoped.
func registerResourceEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ResourceSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ResourceSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewResourceCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewResourceCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewResourceCrudApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p resourceEngineParam) itRes.ResourceCrudRepository {
			return p.Engine.Repository().(itRes.ResourceCrudRepository)
		},
		func(p resourceEngineParam) itRes.ResourceCrudDomainService {
			return p.Engine.DomainService().(itRes.ResourceCrudDomainService)
		},
		func(p resourceEngineParam) itRes.ResourceCrudApplicationService {
			return p.Engine.ApplicationService().(itRes.ResourceCrudApplicationService)
		},
	))
}
