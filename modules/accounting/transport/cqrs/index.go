package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	essconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
)

func InitCqrsHandlers() error {
	return initUsageCheckHandlers()
}

// initUsageCheckHandlers subscribes Accounting's answer to usage checks.
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

		return usagecheck.SubscribeHandler(
			context.Background(), bus, modconstants.AccountingModuleName, registry)
	})
}
