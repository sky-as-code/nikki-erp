package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// InitApplicationServices registers the Sales application services in the DI container.
func InitApplicationServices() error {
	return errors.Join(
		deps.Register(NewSalesChannelApplicationServiceImpl),
		deps.Register(NewSalesPointApplicationServiceImpl),
		deps.Register(NewChannelPaymentApplicationServiceImpl),
		deps.Register(NewPointPaymentApplicationServiceImpl),

		// The in-process selling port. Registered here rather than in infra/external because it is
		// an application service: it authorizes, and every caller of it is subject to that check.
		deps.Register(NewSalesOrderExtService),

		// The pricing rule set another module reads to price its own baskets. Unlike the port above
		// it asserts nothing and is therefore a domain service, registered from here because this
		// is where Sales publishes what other modules may reach.
		deps.Register(services.NewPricingRuleSetExtService),

		deps.Register(services.NewPointPaymentMethodExtService),
	)
}
