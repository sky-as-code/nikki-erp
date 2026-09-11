package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
)

// NewUomRepository is handed the composable default by the UoM onion.
func NewUomRepository(base composable.CrudRepository) itUom.UomRepository {
	return &UomRepositoryImpl{CrudRepository: base}
}

type UomRepositoryImpl struct {
	composable.CrudRepository
}
