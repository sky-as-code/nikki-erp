package app

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/services"
)

// InitApplicationServices registers Accounting's application services.
//
// The thirteen tax resources are served entirely by the dynamic resource engine and need no
// application service of their own. What is registered here is what the engine cannot express:
// calculate, simulate and reverse, which read configuration across several resources and drive the
// arithmetic in domain/services.
func InitApplicationServices() error {
	return stdErr.Join(
		deps.Register(services.NewTaxCalculationDomainServiceImpl),
		deps.Register(NewTaxCalculationApplicationServiceImpl),
	)
}
