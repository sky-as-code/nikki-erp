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

		// The org currency reader is published as its domain service, with no application
		// service wrapping it. Every other capability here has one because it needs an
		// authorization check; this one does not. It reads a setting the caller is already
		// entitled to read — the settings module decides that — and adding a permission of its
		// own would mean a caller could hold the right to read the settings and still be refused
		// the currency they name.
		deps.Register(services.NewOrgCurrencyDomainServiceImpl),
	)
}
