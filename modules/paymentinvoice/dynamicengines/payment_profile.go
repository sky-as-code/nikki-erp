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
	itPaymentProfile "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentprofile"
)

// The payment profile: the merchant credentials a payment settles into. Its domain service
// encrypts "config" on every write and decrypts it on every read.

type paymentProfileEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_paymentinvoice_payment_profile"`
}

func registerPaymentProfileEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.PaymentProfileSchemaName),
		func(
			param composable.BuildParam,
			profiles *services.PaymentProfileDomainService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.PaymentProfileSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewPaymentProfileRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewPaymentProfileCrudDomainService(base, profiles)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewPaymentProfileApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p paymentProfileEngineParam) itPaymentProfile.PaymentProfileCrudRepository {
			return p.Engine.Repository().(itPaymentProfile.PaymentProfileCrudRepository)
		},
		func(p paymentProfileEngineParam) itPaymentProfile.PaymentProfileCrudDomainService {
			return p.Engine.DomainService().(itPaymentProfile.PaymentProfileCrudDomainService)
		},
		func(p paymentProfileEngineParam) itPaymentProfile.PaymentProfileCrudApplicationService {
			return p.Engine.ApplicationService().(itPaymentProfile.PaymentProfileCrudApplicationService)
		},
	))
}
