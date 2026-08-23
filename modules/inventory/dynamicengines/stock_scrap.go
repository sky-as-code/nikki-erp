package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// The scrap document and the one operation that executes it.
//
// Unlike the quant, the scrap keeps its CRUD actions: it is a document a user raises, edits and
// may abandon while it is still draft. What the derived service constrains is *when* — a done
// scrap can be neither edited nor deleted, because the movement it generated is permanent.

const PermissionDoScrap = "do_scrap"

const ActionDoScrap = "do_scrap"

const paramScrapId = "id"

func stockScrapEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.StockScrapSchemaName,
		DefineActions: defineStockScrapActions,
	}
}

// defineStockScrapActions exposes Do Scrap as its own action.
//
// It is a POST rather than an update because it is not an edit to the document: it is the event
// that writes off the goods, and it cannot be undone by editing the record afterwards. Its own
// permission, for the same reason the transfer's validate has one — a role that may correct a
// typo in a scrap note should not thereby be able to destroy stock.
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
