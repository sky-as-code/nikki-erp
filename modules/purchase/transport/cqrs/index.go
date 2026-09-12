package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/purchase/constants"
)

func InitCqrsHandlers() error {
	return initUsageCheckHandlers()
}

// initUsageCheckHandlers subscribes Purchase's answer to usage checks. One subscription serves
// every upstream resource Purchase references; the registry routes by (source module, resource).
func initUsageCheckHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		registry := usagecheck.NewCheckerRegistry()

		if err := registry.Register(
			essconstants.EssentialModuleName,
			usagecheck.ResourceUom,
			usagecheck.NewTableChecker(resolveRepository, uomTables()...),
		); err != nil {
			return err
		}

		if err := registry.Register(
			invconstants.InventoryModuleName,
			usagecheck.ResourceProductVariant,
			usagecheck.NewTableChecker(resolveRepository, variantTables()...),
		); err != nil {
			return err
		}

		return usagecheck.SubscribeHandler(
			context.Background(), bus, constants.PurchaseModuleName, registry)
	})
}
