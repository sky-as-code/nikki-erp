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
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

type entitlementEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_entitlement"`
}

// registerEntitlementEngine declares the iam_entitlement onion. org_id is conditional on an
// entitlement - required only when scope=org - and a domain-scoped entitlement legitimately has
// none, so org scoping is withdrawn rather than hiding those rows; permissions are tenant-scoped.
func registerEntitlementEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.EntitlementSchemaName),
		func(param composable.BuildParam, ruleDeps services.EntitlementCrudDeps) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.EntitlementSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewEntitlementCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewEntitlementCrudDomainService(base, ruleDeps)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewEntitlementCrudApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p entitlementEngineParam) itEnt.EntitlementCrudRepository {
			return p.Engine.Repository().(itEnt.EntitlementCrudRepository)
		},
		func(p entitlementEngineParam) itEnt.EntitlementCrudDomainService {
			return p.Engine.DomainService().(itEnt.EntitlementCrudDomainService)
		},
		func(p entitlementEngineParam) itEnt.EntitlementCrudApplicationService {
			return p.Engine.ApplicationService().(itEnt.EntitlementCrudApplicationService)
		},
	))
}
