package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// The scrap document and the operation that executes it. Unlike the quant, the scrap keeps its
// CRUD actions: it is a document a user raises, edits and may abandon while draft. The derived
// service constrains when — a done scrap can be neither edited nor deleted, because the movement
// it generated is permanent.

const PermissionDoScrap = "do_scrap"

const ActionDoScrap = "do_scrap"

const paramScrapId = "id"

func stockScrapEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockScrapSchemaName,
		DefineActions: defineStockScrapActions,
	}
}

// defineStockScrapActions exposes Do Scrap as its own action. A POST rather than an update: it is
// the event that writes off the goods and cannot be undone by editing the record afterwards. It
// carries its own permission so a role that may fix a typo in a scrap note cannot destroy stock.
func defineStockScrapActions(engine drif.DynamicResourceEngine) error {
	return engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  ActionDoScrap,
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/do_scrap",
		Permission:  PermissionDoScrap,
		MainProcess: processDoScrap,
	})
}

func scrapServiceOf(input drif.ProcessInput) (*services.StockScrapDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.StockScrapDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the stock scrap engine is not running the derived scrap service; " +
				"InventoryModule.Init must install it with SetResourceService")
	}
	return service, nil
}

func processDoScrap(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := scrapServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.DoScrap(ctx, readStringField(input.Params, paramScrapId))
	return toMutateActionResult(result, err)
}
