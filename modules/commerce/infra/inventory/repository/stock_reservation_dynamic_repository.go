package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewStockReservationRepository(base composable.CrudRepository) itStock.StockReservationRepository {
	return &StockReservationRepositoryImpl{CrudRepository: base}
}

type StockReservationRepositoryImpl struct {
	composable.CrudRepository
}
