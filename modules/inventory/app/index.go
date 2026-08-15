package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// InitApplicationServices publishes the orchestration layer into the container.
//
// It runs after the domain services are installed on their engines, because it composes two of
// them; the constructor takes the concrete types rather than resolving them, so a missing
// dependency is a compile error rather than a start-up one.
func InitApplicationServices(
	warehouseSvc *services.WarehouseDomainServiceImpl,
	locationSvc *services.InventoryLocationDomainServiceImpl,
) error {
	appSvc := NewWarehouseAppService(warehouseSvc, locationSvc)
	return deps.Register(func() itWarehouse.WarehouseAppService { return appSvc })
}
