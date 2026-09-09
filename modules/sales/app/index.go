package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices registers the Sales application services in the DI container.
func InitApplicationServices() error {
	return errors.Join(
		deps.Register(NewSalesChannelApplicationServiceImpl),
		deps.Register(NewSalesPointApplicationServiceImpl),
		deps.Register(NewChannelPaymentApplicationServiceImpl),

		// The in-process selling port. Registered here rather than in infra/external because it is
		// an application service: it authorizes, and every caller of it is subject to that check.
		deps.Register(NewSalesOrderExtService),
	)
}
