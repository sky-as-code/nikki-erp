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
	itTransaction "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/transaction"
)

// The transaction: one movement of money against an order, written by the payment flow.

type transactionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_transaction"`
}

func registerTransactionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TransactionSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TransactionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTransactionRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTransactionCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
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
