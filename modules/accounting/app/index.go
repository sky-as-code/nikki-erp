package app

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/services"
)

// InitApplicationServices registers what the dynamic resource engine cannot express: calculate,
// simulate and reverse. The thirteen tax resources are served by the engine itself.
func InitApplicationServices() error {
	return stdErr.Join(
		deps.Register(services.NewTaxCalculationDomainServiceImpl),
		deps.Register(NewTaxCalculationApplicationServiceImpl),

		// The org currency reader has no application service because it needs no authorization
		// check of its own: the settings module already gates the setting it reads, and a second
		// permission could refuse a caller entitled to read those settings.
		deps.Register(services.NewOrgCurrencyDomainServiceImpl),
	)
}
