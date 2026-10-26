package cqrs

import (
	stdErr "errors"

	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// UomUsageChecker answers whether a Purchase table still holds a unit of measure. Commerce
// subscribes one usage handler for the whole module and combines the parts' answers.
func UomUsageChecker() usagecheck.ResourceUsageChecker {
	return usagecheck.NewTableChecker(resolveRepository, uomTables()...)
}

// InventoryUsageRegistry binds a checker to every Inventory resource Commerce guards before a
// delete. Purchase and Inventory are parts of one module, so Inventory asks this registry in
// process rather than over the bus. A resource Purchase holds no reference to is still registered,
// with no tables, so the registry states every resource it was asked about.
func InventoryUsageRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	register := func(resource string, tables ...usagecheck.ReferencingTable) error {
		return registry.Register(commconstants.CommerceModuleName, resource,
			usagecheck.NewTableChecker(resolveRepository, tables...))
	}

	err := stdErr.Join(
		register(usagecheck.ResourceProductVariant, variantTables()...),
		register(usagecheck.ResourceProductTemplate, templateTables()...),
		// No Purchase schema names a stock quant: its documents name the variant and the location.
		register(usagecheck.ResourceStockQuant),
	)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
