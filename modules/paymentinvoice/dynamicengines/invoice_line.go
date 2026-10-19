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

// The invoice line, written by the invoice's own operations or edited on a draft.

type invoiceLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_invoice_line"`
}

func registerInvoiceLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.InvoiceLineSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.InvoiceLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewInvoiceLineRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewInvoiceLineCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewInvoiceLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p invoiceLineEngineParam) itInvoice.InvoiceLineCrudRepository {
			return p.Engine.Repository().(itInvoice.InvoiceLineCrudRepository)
		},
		func(p invoiceLineEngineParam) itInvoice.InvoiceLineCrudDomainService {
			return p.Engine.DomainService().(itInvoice.InvoiceLineCrudDomainService)
		},
		func(p invoiceLineEngineParam) itInvoice.InvoiceLineCrudApplicationService {
			return p.Engine.ApplicationService().(itInvoice.InvoiceLineCrudApplicationService)
		},
	))
}
