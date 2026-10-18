package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/purchase/infra/repository"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/agreement"
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
