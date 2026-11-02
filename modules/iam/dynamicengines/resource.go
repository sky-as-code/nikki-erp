package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

type resourceEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_resource"`
}

// registerResourceEngine declares the iam_resource onion: the catalog of what can be authorized,
// shared by the whole tenant, so its permissions are tenant-scoped.
func registerResourceEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ResourceSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ResourceSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewResourceCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewResourceCrudDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
