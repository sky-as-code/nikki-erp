package app

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// NewStorageCategoryApplicationService is handed the composable default by the storage_category onion.
func NewStorageCategoryApplicationService(base composable.CrudApplicationService) itWarehouse.StorageCategoryApplicationService {
	return &StorageCategoryApplicationServiceImpl{CrudApplicationService: base}
}

// StorageCategoryApplicationServiceImpl is the authorized CRUD of the resource.
type StorageCategoryApplicationServiceImpl struct {
	composable.CrudApplicationService
}
