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
	)
}
