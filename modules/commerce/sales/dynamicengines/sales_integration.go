package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itIntegration "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/integration"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The two self-records, both read-only.

type salesManualDiscountEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_manual_discount"`
}

func registerSalesManualDiscountEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesManualDiscountSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesManualDiscountSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesManualDiscountRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesManualDiscountApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesManualDiscountEngineParam) itIntegration.SalesManualDiscountRepository {
			return p.Engine.Repository().(itIntegration.SalesManualDiscountRepository)
		},
		func(p salesManualDiscountEngineParam) itIntegration.SalesManualDiscountApplicationService {
			return p.Engine.ApplicationService().(itIntegration.SalesManualDiscountApplicationService)
		},
	))
}

type salesIntegrationOutboxEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_integration_outbox"`
}

func registerSalesIntegrationOutboxEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesIntegrationOutboxSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesIntegrationOutboxSchemaName,
				CrudActions: readOnlyCrudActions(),
				// Not audited: the outbox is delivery plumbing, and every event it carries describes a
				// write the audit log already recorded.
				IsAudited: util.ToPtr(false),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesIntegrationOutboxRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesIntegrationOutboxApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesIntegrationOutboxEngineParam) itIntegration.SalesIntegrationOutboxRepository {
			return p.Engine.Repository().(itIntegration.SalesIntegrationOutboxRepository)
		},
		func(p salesIntegrationOutboxEngineParam) itIntegration.SalesIntegrationOutboxApplicationService {
			return p.Engine.ApplicationService().(itIntegration.SalesIntegrationOutboxApplicationService)
		},
	))
}
