package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// NewUomCatRepository is handed the composable default by the UoM Category onion.
func NewUomCatRepository(base dynamicresource.CrudRepository) itUomCat.UomCatRepository {
	return &UomCatRepositoryImpl{CrudRepository: base}
}

type UomCatRepositoryImpl struct {
	dynamicresource.CrudRepository
}
