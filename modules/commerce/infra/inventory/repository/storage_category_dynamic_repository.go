package repository

import (
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewStorageCategoryRepository is handed the composable default by the storage_category onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStorageCategoryRepository(base dynamicresource.CrudRepository) itWarehouse.StorageCategoryRepository {
	return &StorageCategoryRepositoryImpl{CrudRepository: base}
}

type StorageCategoryRepositoryImpl struct {
	dynamicresource.CrudRepository
}
