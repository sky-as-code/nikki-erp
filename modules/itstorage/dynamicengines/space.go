package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/itstorage/infra/repository"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

type spaceEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_itstorage_space"`
}

// registerSpaceEngine keeps every built-in action in the onion: the org lifecycle creates spaces
// and a rename updates one through the domain service. Only the reads get routes.
func registerSpaceEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SpaceSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SpaceSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSpaceRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSpaceDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSpaceApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p spaceEngineParam) itSpace.SpaceRepository {
			return p.Engine.Repository().(itSpace.SpaceRepository)
		},
		func(p spaceEngineParam) itSpace.SpaceDomainService {
			return p.Engine.DomainService().(itSpace.SpaceDomainService)
		},
		func(p spaceEngineParam) itSpace.SpaceApplicationService {
			return p.Engine.ApplicationService().(itSpace.SpaceApplicationService)
		},
	))
}
