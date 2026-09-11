package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewStorageCategoryApplicationService is handed the composable default by the storage_category onion.
func NewStorageCategoryApplicationService(base composable.CrudApplicationService) itWarehouse.StorageCategoryApplicationService {
	return &StorageCategoryApplicationServiceImpl{CrudApplicationService: base}
}

// StorageCategoryApplicationServiceImpl is the authorized CRUD of the resource.
type StorageCategoryApplicationServiceImpl struct {
	composable.CrudApplicationService
}
