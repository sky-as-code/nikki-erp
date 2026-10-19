package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

type groupEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_iam_group"`
}

// registerGroupEngine declares the iam_group onion. The schema has no org_id, so the default
// org scoping leaves it alone and permissions are checked without an org.
func registerGroupEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.GroupSchemaName),
		func(
			param composable.BuildParam,
			permRepo itPerm.PermissionRepository, historyRepo itPerm.PermissionHistoryRepository,
		) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.GroupSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewGroupCrudRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewGroupCrudDomainService(base, permRepo, historyRepo)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewGroupCrudApplicationService(base, base.DomainService().(itGroup.GroupCrudDomainService))
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p groupEngineParam) itGroup.GroupCrudRepository {
			return p.Engine.Repository().(itGroup.GroupCrudRepository)
		},
		func(p groupEngineParam) itGroup.GroupCrudDomainService {
			return p.Engine.DomainService().(itGroup.GroupCrudDomainService)
		},
		func(p groupEngineParam) itGroup.GroupCrudApplicationService {
			return p.Engine.ApplicationService().(itGroup.GroupCrudApplicationService)
		},
	))
}
