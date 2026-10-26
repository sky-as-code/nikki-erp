package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ConsumptionEngineName is the container name of the onion, checked against
// composable.EngineDependencyName by a test.
const ConsumptionEngineName = "dynengine_itstorage_consumption"

type consumptionRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_itstorage_consumption"`
}

func NewConsumptionRest(params consumptionRestParams) *ConsumptionRest {
	rest := &ConsumptionRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// ConsumptionRest serves the read-only built-ins: the counters are kept by the file paths.
type ConsumptionRest struct {
	composable.CrudRestBase
}
