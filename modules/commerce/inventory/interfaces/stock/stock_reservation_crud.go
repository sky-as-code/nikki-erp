package stock

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The warehouse-level reservation resource. Its rows are written only by the reservation
// operations (reserve, consume, release, protect, expire): the built-in create, update and delete
// are withheld at the engine, so the read surface is all a client reaches directly.

type StockReservationRepository interface {
	composable.CrudRepository
}

type StockReservationDomainService interface {
	composable.CrudDomainService
}

type StockReservationApplicationService interface {
	composable.CrudApplicationService
	StockReservationActionService
}

type (
	GetStockReservationByIdQuery = composable.GetByIdQuery
	SearchStockReservationsQuery = composable.SearchQuery
	StockReservationExistsQuery  = composable.ExistsQuery
)

type (
	GetStockReservationByIdResult = composable.GetOneResult
	SearchStockReservationsResult = composable.SearchResult
	StockReservationExistsResult  = composable.ExistsResult
)

// The guard is a lock target with no client surface at all; the outbox is Inventory's own event
// queue. Both are onions only so their repositories exist for the services that use them.

type WarehouseProductGuardRepository interface {
	composable.CrudRepository
}

type InventoryIntegrationOutboxRepository interface {
	composable.CrudRepository
}
