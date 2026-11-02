package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The warehouse-level reservation resource. Its rows are written only by the reservation
// operations (reserve, consume, release, protect, expire): the built-in create, update and delete
// are withheld at the engine, so the read surface is all a client reaches directly.

type StockReservationRepository interface {
	dynamicresource.CrudRepository
}

type StockReservationDomainService interface {
	dynamicresource.CrudDomainService
}

type StockReservationApplicationService interface {
	dynamicresource.CrudApplicationService
	StockReservationActionService
}

type (
	GetStockReservationByIdQuery = dynamicresource.GetByIdQuery
	SearchStockReservationsQuery = dynamicresource.SearchQuery
	StockReservationExistsQuery  = dynamicresource.ExistsQuery
)

type (
	GetStockReservationByIdResult = dynamicresource.GetOneResult
	SearchStockReservationsResult = dynamicresource.SearchResult
	StockReservationExistsResult  = dynamicresource.ExistsResult
)

// The guard is a lock target with no client surface at all; the outbox is Inventory's own event
// queue. Both are onions only so their repositories exist for the services that use them.

type WarehouseProductGuardRepository interface {
	dynamicresource.CrudRepository
}

type InventoryIntegrationOutboxRepository interface {
	dynamicresource.CrudRepository
}
