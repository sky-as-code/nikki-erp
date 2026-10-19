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
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

type roleEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_role"`
}

// registerRoleEngine declares the iam_role onion. org_id is optional on a role: a role with no org
// accepts only domain-scoped entitlements and is usable across organizations, so requiring org_id
// would make every such role invisible, including during role assignment. Org scoping is
// withdrawn and permissions are tenant-scoped.
func registerRoleEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.RoleSchemaName),
		func(param composable.BuildParam, ruleDeps services.RoleCrudDeps) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:      models.RoleSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewRoleCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewRoleCrudDomainService(base, ruleDeps)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewRoleCrudApplicationService(base, base.DomainService().(itRole.RoleCrudDomainService))
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p roleEngineParam) itRole.RoleCrudRepository {
			return p.Engine.Repository().(itRole.RoleCrudRepository)
		},
		func(p roleEngineParam) itRole.RoleCrudDomainService {
			return p.Engine.DomainService().(itRole.RoleCrudDomainService)
		},
		func(p roleEngineParam) itRole.RoleCrudApplicationService {
			return p.Engine.ApplicationService().(itRole.RoleCrudApplicationService)
		},
	))
}
