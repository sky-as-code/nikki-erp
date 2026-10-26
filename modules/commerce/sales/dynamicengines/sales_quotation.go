package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itQuotation "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/quotation"
)

type salesQuotationEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_quotation"`
}

func registerSalesQuotationEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesQuotationSchemaName),
		func(
			param composable.BuildParam,
			settings itExt.EffectiveSettingsExtService,
			tax itExt.TaxCalculationExtService,
			products itExt.ProductVariantExtService,
			basis itExt.ProductPricingBasisExtService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesQuotationSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesQuotationRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesQuotationApplicationService(base, settings, tax, products, basis)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesQuotationEngineParam) itQuotation.SalesQuotationRepository {
			return p.Engine.Repository().(itQuotation.SalesQuotationRepository)
		},
		func(p salesQuotationEngineParam) itQuotation.SalesQuotationApplicationService {
			return p.Engine.ApplicationService().(itQuotation.SalesQuotationApplicationService)
		},
	))
}

type salesQuotationLineEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_quotation_line"`
}

func registerSalesQuotationLineEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesQuotationLineSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesQuotationLineSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesQuotationLineRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesQuotationLineApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesQuotationLineEngineParam) itQuotation.SalesQuotationLineRepository {
			return p.Engine.Repository().(itQuotation.SalesQuotationLineRepository)
		},
		func(p salesQuotationLineEngineParam) itQuotation.SalesQuotationLineApplicationService {
			return p.Engine.ApplicationService().(itQuotation.SalesQuotationLineApplicationService)
		},
	))
}
