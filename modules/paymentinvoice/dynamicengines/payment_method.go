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
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
	itPaymentMethod "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

// The payment method. The method list is data rather than an enum, so this resource is how a
// deployment adds a gateway account, withdraws one, or adjusts the amounts a gateway will accept.
// Its application service also answers the PaymentMethodAppService port, which needs the gateway
// registry to judge whether this build ships the adapter a method names.

type paymentMethodEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_payment_method"`
}

func registerPaymentMethodEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.PaymentMethodSchemaName),
		func(
			param composable.BuildParam,
			registry *itGateway.Registry,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.PaymentMethodSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewPaymentMethodRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewPaymentMethodCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewPaymentMethodApplicationService(base, registry)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p paymentMethodEngineParam) itPaymentMethod.PaymentMethodCrudRepository {
			return p.Engine.Repository().(itPaymentMethod.PaymentMethodCrudRepository)
		},
		func(p paymentMethodEngineParam) itPaymentMethod.PaymentMethodCrudDomainService {
			return p.Engine.DomainService().(itPaymentMethod.PaymentMethodCrudDomainService)
		},
		func(p paymentMethodEngineParam) itPaymentMethod.PaymentMethodCrudApplicationService {
			return p.Engine.ApplicationService().(itPaymentMethod.PaymentMethodCrudApplicationService)
		},
		func(p paymentMethodEngineParam) itPaymentMethod.PaymentMethodAppService {
			return p.Engine.ApplicationService().(itPaymentMethod.PaymentMethodAppService)
		},
	))
}
