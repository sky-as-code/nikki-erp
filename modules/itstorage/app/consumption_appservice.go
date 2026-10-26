package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

// NewConsumptionApplicationService is handed the composable default by the itstorage_consumption
// onion. The counters are kept by the file paths, so only the authorized reads have routes.
func NewConsumptionApplicationService(base composable.CrudApplicationService) itConsumption.ConsumptionApplicationService {
	return &ConsumptionApplicationServiceImpl{CrudApplicationService: base}
}

type ConsumptionApplicationServiceImpl struct {
	composable.CrudApplicationService
}
