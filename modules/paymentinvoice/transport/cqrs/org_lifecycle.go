package cqrs

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

// InitCqrsHandlers subscribes this module to the org lifecycle commands iam sends and to the usage
// checks Essential sends.
func InitCqrsHandlers() error {
	if err := initUsageCheckHandlers(); err != nil {
		return err
	}
	return initOrgLifecycleHandlers()
}

// initOrgLifecycleHandlers builds the lifecycle lazily: the numbering it registers the invoice
// series through is resolved on the first command, when every port is guaranteed to be built.
func initOrgLifecycleHandlers() error {
	svc := orglifecycle.Lazy(func() (orglifecycle.OrgLifecycleService, error) {
		var built orglifecycle.OrgLifecycleService
		err := deps.Invoke(func(sequences itExt.SequenceExtService, seeder itExt.OrgSettingsSeedExtService) {
			built = app.NewOrgLifecycleAppService(sequences, seeder)
		})
		return built, err
	})
	return deps.Invoke(func(bus cqrs.CqrsBus) error {
		return orglifecycle.RegisterHandlers(context.Background(), bus, modconstants.PaymentInvoiceModuleName, svc)
	})
}
