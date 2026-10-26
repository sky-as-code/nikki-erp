package v1

import (
	"go.uber.org/dig"

	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// StorageCategoryEngineName is the container name of the storage_category onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const StorageCategoryEngineName = "dynengine_inventory_storage_category"

type storageCategoryRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_storage_category"`
}

func NewStorageCategoryRest(params storageCategoryRestParams) *StorageCategoryRest {
	rest := &StorageCategoryRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.storageCategorySvc = params.Engine.ApplicationService().(itWarehouse.StorageCategoryApplicationService)
	return rest
}

// StorageCategoryRest serves the built-in CRUD of the storage_category resource.
type StorageCategoryRest struct {
	composable.CrudRestBase
	storageCategorySvc itWarehouse.StorageCategoryApplicationService
}
