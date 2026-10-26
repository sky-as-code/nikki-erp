package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type storageCategoryEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_storage_category"`
}

// registerStorageCategoryEngine declares the storage_category onion and publishes its typed layers.
func registerStorageCategoryEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.StorageCategorySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.StorageCategorySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewStorageCategoryRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewStorageCategoryDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewStorageCategoryApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p storageCategoryEngineParam) itWarehouse.StorageCategoryRepository {
			return p.Engine.Repository().(itWarehouse.StorageCategoryRepository)
		},
		func(p storageCategoryEngineParam) itWarehouse.StorageCategoryDomainService {
			return p.Engine.DomainService().(itWarehouse.StorageCategoryDomainService)
		},
		func(p storageCategoryEngineParam) itWarehouse.StorageCategoryApplicationService {
			return p.Engine.ApplicationService().(itWarehouse.StorageCategoryApplicationService)
		},
	))
}
