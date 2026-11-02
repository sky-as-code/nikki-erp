package repository

import (
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewStockReservationRepository(base dynamicresource.CrudRepository) itStock.StockReservationRepository {
	return &StockReservationRepositoryImpl{CrudRepository: base}
}

type StockReservationRepositoryImpl struct {
	dynamicresource.CrudRepository
}
