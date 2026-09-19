package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

func NewStockReservationRepository(base composable.CrudRepository) itStock.StockReservationRepository {
	return &StockReservationRepositoryImpl{CrudRepository: base}
}

type StockReservationRepositoryImpl struct {
	composable.CrudRepository
}
