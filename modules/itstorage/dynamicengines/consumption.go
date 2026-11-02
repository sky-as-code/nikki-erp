package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/app"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/itstorage/infra/repository"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

type consumptionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_itstorage_consumption"`
}

// registerConsumptionEngine needs no org scoping flag: the schema has no org_id column, which the
// engine skips by itself. A row belongs to its owner by owner_type and owner_id. Only reads have
// routes.
func registerConsumptionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.ConsumptionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.ConsumptionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewConsumptionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewConsumptionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewConsumptionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p consumptionEngineParam) itConsumption.ConsumptionRepository {
			return p.Engine.Repository().(itConsumption.ConsumptionRepository)
		},
		func(p consumptionEngineParam) itConsumption.ConsumptionDomainService {
			return p.Engine.DomainService().(itConsumption.ConsumptionDomainService)
		},
		func(p consumptionEngineParam) itConsumption.ConsumptionApplicationService {
			return p.Engine.ApplicationService().(itConsumption.ConsumptionApplicationService)
		},
	))
}
