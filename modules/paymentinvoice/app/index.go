package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices publishes this module's outward-facing application services into the
// container, so that another module's adapter can resolve them by interface.
//
// It must run after the gateway registry is registered, because the payment-method service is
// constructed with it: one of the usability gates asks the registry whether this build ships the
// adapter a method names.
func InitApplicationServices() error {
	return errors.Join(
		deps.Register(NewPaymentMethodApplicationServiceImpl),
	)
}
