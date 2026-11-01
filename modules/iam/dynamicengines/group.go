package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/app"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/iam/infra/repository"
	itGroup "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/group"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

type groupEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_group"`
}

// registerGroupEngine declares the iam_group onion. The schema has no org_id, so the default
// org scoping leaves it alone and permissions are checked without an org.
func registerGroupEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.GroupSchemaName),
		func(
			param dynamicresource.BuildParam,
			permRepo itPerm.PermissionRepository, historyRepo itPerm.PermissionHistoryRepository,
			userRepo itUser.UserCrudRepository,
		) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.GroupSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewGroupCrudRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewGroupCrudDomainService(base, permRepo, historyRepo, userRepo)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
