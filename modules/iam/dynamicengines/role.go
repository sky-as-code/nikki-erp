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
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

type roleEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_role"`
}

// registerRoleEngine declares the iam_role onion. org_id is optional on a role: a role with no org
// accepts only domain-scoped entitlements and is usable across organizations, so requiring org_id
// would make every such role invisible, including during role assignment. Org scoping is
// withdrawn and permissions are tenant-scoped.
func registerRoleEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.RoleSchemaName),
		func(param dynamicresource.BuildParam, ruleDeps services.RoleCrudDeps) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.RoleSchemaName,
				IsOrgScoped:     util.ToPtr(false),
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewRoleCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewRoleCrudDomainService(base, ruleDeps)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
