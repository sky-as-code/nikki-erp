package services

import (
	"context"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Inventory's answer to "is this unit of measure still in use", asked before Essential deletes
// or re-scales one.
//
// Inventory is the heaviest consumer of units in the application and registered no probe at all
// until now: a unit referenced only by stock could be edited freely, which is precisely the case
// BR-UOM-ESS-020 exists to prevent. Both kinds of reference count and for the same reason -
// a movement records how much was moved, a configuration says what the balance is counted in,
// and neither survives losing the unit.

func uomTables() []usagecheck.ReferencingTable {
	return []usagecheck.ReferencingTable{
		{SchemaName: models.StockMoveSchemaName, Field: models.StockMoveFieldUomId},
		{SchemaName: models.StockMoveLineSchemaName, Field: models.StockMoveLineFieldUomId},
		{SchemaName: models.StockQuantSchemaName, Field: models.StockQuantFieldBaseUomId},
		{SchemaName: models.StockScrapSchemaName, Field: models.StockScrapFieldBaseUomId},
		{SchemaName: models.StockProductConfigSchemaName, Field: models.StockProductConfigFieldInventoryUomId},
	}
}

// resolveRepository reaches Inventory's own repositories through the module-private hub, at
// check time rather than at registration: the hub is filled while the engines are built, which
// happens after the checker is registered.
func resolveRepository(schemaName string) (usagecheck.RowSearcher, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "no repository for %s", schemaName)
	}
	return repo, nil
}

// RegisterUsageCheckers subscribes Inventory's answer to usage checks.
func RegisterUsageCheckers(bus cqrs.CqrsBus) error {
	registry := usagecheck.NewCheckerRegistry()

	if err := registry.Register(
		essconstants.EssentialModuleName,
		usagecheck.ResourceUom,
		usagecheck.NewTableChecker(resolveRepository, uomTables()...),
	); err != nil {
		return err
	}

	return usagecheck.SubscribeHandler(
		context.Background(), bus, modconstants.InventoryModuleName, registry)
}
