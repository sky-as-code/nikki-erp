package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/purchase/infra/repository"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

// The sourcing group, created and reaped by the order alternatives; client create and delete are
// refused by its domain service.

type sourcingGroupEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_sourcing_group"`
}

func registerSourcingGroupEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SourcingGroupSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SourcingGroupSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSourcingGroupRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewSourcingGroupCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSourcingGroupApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p sourcingGroupEngineParam) itOrder.SourcingGroupCrudRepository {
			return p.Engine.Repository().(itOrder.SourcingGroupCrudRepository)
		},
		func(p sourcingGroupEngineParam) itOrder.SourcingGroupCrudDomainService {
			return p.Engine.DomainService().(itOrder.SourcingGroupCrudDomainService)
		},
		func(p sourcingGroupEngineParam) itOrder.SourcingGroupCrudApplicationService {
			return p.Engine.ApplicationService().(itOrder.SourcingGroupCrudApplicationService)
		},
	))
}
