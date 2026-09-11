package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewStorageCategoryRepository is handed the composable default by the storage_category onion. The type exists so a
// bespoke query has a home; it adds nothing today.
func NewStorageCategoryRepository(base composable.CrudRepository) itWarehouse.StorageCategoryRepository {
	return &StorageCategoryRepositoryImpl{CrudRepository: base}
}

type StorageCategoryRepositoryImpl struct {
	composable.CrudRepository
}
