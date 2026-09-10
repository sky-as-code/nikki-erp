package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// NewUomCatRepository is handed the composable default by the UoM Category onion.
func NewUomCatRepository(base composable.CrudRepository) itUomCat.UomCatRepository {
	return &UomCatRepositoryImpl{CrudRepository: base}
}

type UomCatRepositoryImpl struct {
	composable.CrudRepository
}
