package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/purchase/constants"
)

// initOrgLifecycleHandlers subscribes this module to the org lifecycle commands iam sends.
func initOrgLifecycleHandlers() error {
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, modconstants.PurchaseModuleName,
			app.NewOrgLifecycleAppService())
	})
}
