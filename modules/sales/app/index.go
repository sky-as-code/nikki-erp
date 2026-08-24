package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices publishes the Sales application services into the container, so that a
// transport handler or another module's adapter can resolve them by interface.
func InitApplicationServices() error {
	return errors.Join(
		deps.Register(NewSalesChannelApplicationServiceImpl),
		deps.Register(NewSalesPointApplicationServiceImpl),
	)
}
