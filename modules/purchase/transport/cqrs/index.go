package cqrs

import (
	"context"
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/purchase/constants"
)

func InitCqrsHandlers() error {
	if err := initUsageCheckHandlers(); err != nil {
		return err
	}
	return initOrgLifecycleHandlers()
}

// initUsageCheckHandlers subscribes Purchase's answer to usage checks. One subscription serves
// every upstream resource Purchase references; the registry routes by (source module, resource).
func initUsageCheckHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		registry, err := newUsageCheckRegistry()
		if err != nil {
			return err
		}
		return usagecheck.SubscribeHandler(
			context.Background(), bus, constants.PurchaseModuleName, registry)
	})
}

// newUsageCheckRegistry binds a checker to every (owner, resource) pair the owners dispatch to
// Purchase. Inventory asks each of its dependants about every resource it guards, so a pair Purchase
// holds no reference to is still registered, with no tables, to answer "unused": an unregistered
// resource is answered with an error, and an unanswered check refuses the delete.
func newUsageCheckRegistry() (*usagecheck.CheckerRegistry, error) {
	registry := usagecheck.NewCheckerRegistry()
	register := func(owner, resource string, tables ...usagecheck.ReferencingTable) error {
		return registry.Register(owner, resource, usagecheck.NewTableChecker(resolveRepository, tables...))
	}

	err := stdErr.Join(
		register(essconstants.EssentialModuleName, usagecheck.ResourceUom, uomTables()...),
		register(invconstants.InventoryModuleName, usagecheck.ResourceProductVariant, variantTables()...),
		register(invconstants.InventoryModuleName, usagecheck.ResourceProductTemplate, templateTables()...),
		// No Purchase schema names a stock quant: its documents name the variant and the location.
		register(invconstants.InventoryModuleName, usagecheck.ResourceStockQuant),
	)
	if err != nil {
		return nil, err
	}
	return registry, nil
}
