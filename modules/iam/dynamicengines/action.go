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
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

type actionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_action"`
}

// registerActionEngine declares the iam_action onion, tenant-scoped like the resource catalog it
// belongs to. A client lists one resource's actions by filtering on resource_id.
func registerActionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.ActionSchemaName),
		func(param composable.BuildParam, resourceRepo itRes.ResourceCrudRepository) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ActionSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewActionCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewActionCrudDomainService(base, resourceRepo)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewActionCrudApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p actionEngineParam) itAct.ActionCrudRepository {
			return p.Engine.Repository().(itAct.ActionCrudRepository)
		},
		func(p actionEngineParam) itAct.ActionCrudDomainService {
			return p.Engine.DomainService().(itAct.ActionCrudDomainService)
		},
		func(p actionEngineParam) itAct.ActionCrudApplicationService {
			return p.Engine.ApplicationService().(itAct.ActionCrudApplicationService)
		},
	))
}
