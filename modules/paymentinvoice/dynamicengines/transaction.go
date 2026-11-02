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
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

// The transaction: one movement of money against an order, written by the payment flow.

type transactionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_transaction"`
}

func registerTransactionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.TransactionSchemaName),
		func(
			param dynamicresource.BuildParam,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.TransactionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewTransactionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewTransactionCrudDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewTransactionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p transactionEngineParam) itTransaction.TransactionCrudRepository {
			return p.Engine.Repository().(itTransaction.TransactionCrudRepository)
		},
		func(p transactionEngineParam) itTransaction.TransactionCrudDomainService {
			return p.Engine.DomainService().(itTransaction.TransactionCrudDomainService)
		},
		func(p transactionEngineParam) itTransaction.TransactionCrudApplicationService {
			return p.Engine.ApplicationService().(itTransaction.TransactionCrudApplicationService)
		},
	))
}
