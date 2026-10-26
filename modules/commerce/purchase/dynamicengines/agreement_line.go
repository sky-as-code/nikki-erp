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
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The agreement line.

type agreementLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_purchase_agreement_line"`
}

func registerAgreementLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.AgreementLineSchemaName),
		func(
			param composable.BuildParam,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.AgreementLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewAgreementLineRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewAgreementLineCrudDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewAgreementLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p agreementLineEngineParam) itAgreement.AgreementLineCrudRepository {
			return p.Engine.Repository().(itAgreement.AgreementLineCrudRepository)
		},
		func(p agreementLineEngineParam) itAgreement.AgreementLineCrudDomainService {
			return p.Engine.DomainService().(itAgreement.AgreementLineCrudDomainService)
		},
		func(p agreementLineEngineParam) itAgreement.AgreementLineCrudApplicationService {
			return p.Engine.ApplicationService().(itAgreement.AgreementLineCrudApplicationService)
		},
	))
}
