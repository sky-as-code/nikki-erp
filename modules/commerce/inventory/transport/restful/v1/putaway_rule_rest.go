package v1

import (
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// PutawayRuleEngineName is the container name of the putaway_rule onion, declared once for the struct tag and
// checked against composable.EngineDependencyName by a test.
const PutawayRuleEngineName = "dynengine_inventory_putaway_rule"

type putawayRuleRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_inventory_putaway_rule"`
}

func NewPutawayRuleRest(params putawayRuleRestParams) *PutawayRuleRest {
	rest := &PutawayRuleRest{}
	rest.SetApplicationService(params.Engine.ApplicationService())
	rest.putawayRuleSvc = params.Engine.ApplicationService().(itWarehouse.PutawayRuleApplicationService)
	return rest
}

// PutawayRuleRest serves the built-in CRUD of the putaway_rule resource.
type PutawayRuleRest struct {
	composable.CrudRestBase
	putawayRuleSvc itWarehouse.PutawayRuleApplicationService
}
