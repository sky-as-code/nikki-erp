package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/purchase/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The purchase agreement, with confirm, close, cancel and create_rfq.

type agreementEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_purchase_agreement"`
}

func registerAgreementEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.AgreementSchemaName),
		func(
			param dynamicresource.BuildParam,
			vendors itExt.VendorExtService,
			currencies itExt.CurrencyExtService,
			numbering *services.CodeNumbering,
		) dynamicresource.DynamicResourceEngineOnion {
			references := services.NewOrderReferenceValidator(vendors, currencies)
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.AgreementSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewAgreementRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewPurchaseAgreementDomainService(base, references, numbering)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewAgreementApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p agreementEngineParam) itAgreement.AgreementCrudRepository {
			return p.Engine.Repository().(itAgreement.AgreementCrudRepository)
		},
		func(p agreementEngineParam) itAgreement.AgreementCrudDomainService {
			return p.Engine.DomainService().(itAgreement.AgreementCrudDomainService)
		},
		func(p agreementEngineParam) itAgreement.AgreementCrudApplicationService {
			return p.Engine.ApplicationService().(itAgreement.AgreementCrudApplicationService)
		},
	))
}
