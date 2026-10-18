package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	itOrg "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/organization"
)

type orgEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_org"`
}

// registerOrgEngine declares the iam_org onion. An organization is the org scope itself, so its
// permissions are tenant-scoped and the schema carries no org_id to confine by.
func registerOrgEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.OrganizationSchemaName),
		func(
			param composable.BuildParam,
			settingsSvc itExt.OrgSettingsInitExtService, usageDispatcher *usagecheck.Dispatcher,
		) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.OrganizationSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewOrgCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewOrgCrudDomainService(base, settingsSvc, usageDispatcher)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewOrgCrudApplicationService(base, base.DomainService().(itOrg.OrgCrudDomainService))
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p orgEngineParam) itOrg.OrgCrudRepository {
			return p.Engine.Repository().(itOrg.OrgCrudRepository)
		},
		func(p orgEngineParam) itOrg.OrgCrudDomainService {
			return p.Engine.DomainService().(itOrg.OrgCrudDomainService)
		},
		func(p orgEngineParam) itOrg.OrgCrudApplicationService {
			return p.Engine.ApplicationService().(itOrg.OrgCrudApplicationService)
		},
	))
}
