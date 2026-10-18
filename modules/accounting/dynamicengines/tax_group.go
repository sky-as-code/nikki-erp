package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/accounting/app"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/services"
	repo "github.com/sky-as-code/nikki-erp/modules/accounting/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/taxconfig"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type taxGroupEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_accounting_tax_group"`
}

func registerTaxGroupEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.TaxGroupSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return composable.MustBuild(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.TaxGroupSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewTaxGroupRepository(base)
				},
				NewDomainServiceFn: func(base composable.CrudDomainService) composable.CrudDomainService {
					return services.NewTaxGroupDomainService(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewTaxGroupApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p taxGroupEngineParam) taxconfig.TaxGroupRepository {
			return p.Engine.Repository().(taxconfig.TaxGroupRepository)
		},
		func(p taxGroupEngineParam) taxconfig.TaxGroupDomainService {
			return p.Engine.DomainService().(taxconfig.TaxGroupDomainService)
		},
		func(p taxGroupEngineParam) taxconfig.TaxGroupApplicationService {
			return p.Engine.ApplicationService().(taxconfig.TaxGroupApplicationService)
		},
	))
}
