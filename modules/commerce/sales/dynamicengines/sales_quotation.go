package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	itQuotation "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/quotation"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type salesQuotationEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_quotation"`
}

func registerSalesQuotationEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesQuotationSchemaName),
		func(
			param dynamicresource.BuildParam,
			settings itExt.EffectiveSettingsExtService,
			tax itExt.TaxCalculationExtService,
			products itExt.ProductVariantExtService,
			basis itExt.ProductPricingBasisExtService,
		) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesQuotationSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesQuotationRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_quotation_line"`
}

func registerSalesQuotationLineEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesQuotationLineSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesQuotationLineSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesQuotationLineRepository(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
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
