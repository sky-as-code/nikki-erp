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
	itEnt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/entitlement"
)

type entitlementEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_entitlement"`
}

// registerEntitlementEngine declares the iam_entitlement onion. org_id is conditional on an
// entitlement - required only when scope=org - and a domain-scoped entitlement legitimately has
// none, so org scoping is withdrawn rather than hiding those rows; permissions are tenant-scoped.
func registerEntitlementEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.EntitlementSchemaName),
		func(
			param dynamicresource.BuildParam, ruleDeps services.EntitlementCrudDeps,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.EntitlementSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewEntitlementCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewEntitlementCrudDomainService(base, ruleDeps)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
