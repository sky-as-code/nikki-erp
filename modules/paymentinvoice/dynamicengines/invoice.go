package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/repository"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

// The invoice, with the issue action, which delegates to the invoice domain service.

type invoiceEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_invoice"`
}

func registerInvoiceEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.InvoiceSchemaName),
		func(
			param dynamicresource.BuildParam,
			invoices *services.InvoiceDomainService,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.InvoiceSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewInvoiceRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewInvoiceCrudDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewInvoiceApplicationService(base, invoices)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p invoiceEngineParam) itInvoice.InvoiceCrudRepository {
			return p.Engine.Repository().(itInvoice.InvoiceCrudRepository)
		},
		func(p invoiceEngineParam) itInvoice.InvoiceCrudDomainService {
			return p.Engine.DomainService().(itInvoice.InvoiceCrudDomainService)
		},
		func(p invoiceEngineParam) itInvoice.InvoiceCrudApplicationService {
			return p.Engine.ApplicationService().(itInvoice.InvoiceCrudApplicationService)
		},
	))
}
