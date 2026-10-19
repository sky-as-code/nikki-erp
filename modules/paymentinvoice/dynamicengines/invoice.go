package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/app"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/repository"
	itInvoice "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/invoice"
)

// The invoice, with the issue action, which delegates to the invoice domain service.

type invoiceEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_invoice"`
}

func registerInvoiceEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.InvoiceSchemaName),
		func(
			param composable.BuildParam,
			invoices *services.InvoiceDomainService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.InvoiceSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewInvoiceRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewInvoiceCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
