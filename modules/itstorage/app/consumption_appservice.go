package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

// NewConsumptionApplicationService is handed the composable default by the itstorage_consumption
// onion. The counters are kept by the file paths, so only the authorized reads have routes.
func NewConsumptionApplicationService(
	base dynamicresource.CrudApplicationService,
) itConsumption.ConsumptionApplicationService {
	return &ConsumptionApplicationServiceImpl{CrudApplicationService: base}
}

type ConsumptionApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
