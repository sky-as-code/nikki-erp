package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// NewDeliveryRepository is handed the composable default by the delivery onion. Deliveries are
// written and read one row at a time, so the generated CRUD is the whole of it.
func NewDeliveryRepository(base dynamicresource.CrudRepository) it.DeliveryRepository {
	return &DeliveryRepositoryImpl{CrudRepository: base}
}

type DeliveryRepositoryImpl struct {
	dynamicresource.CrudRepository
}
