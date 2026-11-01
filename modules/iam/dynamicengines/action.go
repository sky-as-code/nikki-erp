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
	itAct "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/action"
	itRes "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/resource"
)

type actionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_action"`
}

// registerActionEngine declares the iam_action onion, tenant-scoped like the resource catalog it
// belongs to. A client lists one resource's actions by filtering on resource_id.
func registerActionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ActionSchemaName),
		func(
			param dynamicresource.BuildParam, resourceRepo itRes.ResourceCrudRepository,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.ActionSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewActionCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewActionCrudDomainService(base, resourceRepo)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
