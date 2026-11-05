package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	repo "github.com/sky-as-code/nikki-erp/modules/commerce/infra/sales/repository"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The promotion and voucher resources (BR §8). The promotion, its rules and its rewards are
// editable configuration; the selectors and the confirmed-order records are read-only to clients.

func readOnlyAppService(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
	return app.NewSalesPromotionReadOnlyApplicationService(base)
}

func childRepository(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
	return repo.NewSalesPromotionChildRepository(base)
}

type salesPromotionEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_sales_promotion"`
}

func registerSalesPromotionEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesPromotionSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionSchemaName,
				NewRepositoryFn: func(base dynamicresource.CrudRepository) dynamicresource.CrudRepository {
					return repo.NewSalesPromotionRepository(base)
				},
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSalesPromotionDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesPromotionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionEngineParam) itPromotion.SalesPromotionRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionRepository)
		},
		func(p salesPromotionEngineParam) itPromotion.SalesPromotionDomainService {
			return p.Engine.DomainService().(itPromotion.SalesPromotionDomainService)
		},
		func(p salesPromotionEngineParam) itPromotion.SalesPromotionApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionApplicationService)
		},
	))
}

func registerSalesPromotionRuleEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesPromotionRuleSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.SalesPromotionRuleSchemaName,
				NewRepositoryFn: childRepository,
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSalesPromotionRuleDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesPromotionRuleApplicationService(base)
				},
				ComputedFunctions: map[string]dynamicresource.ComputedFieldFn{
					services.FunctionRuleHasMixedUom: services.RuleHasMixedUomFn,
				},
			}, param)
		},
	)
}

func registerSalesPromotionRewardEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesPromotionRewardSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.SalesPromotionRewardSchemaName,
				NewRepositoryFn: childRepository,
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSalesPromotionRewardDomainService(base)
				},
				NewAppServiceFn: func(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
					return app.NewSalesPromotionRewardApplicationService(base)
				},
			}, param)
		},
	)
}

func registerSalesVoucherEngine() error {
	return deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.SalesVoucherSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.SalesVoucherSchemaName,
				NewRepositoryFn: childRepository,
				NewDomainServiceFn: func(base dynamicresource.CrudDomainService) dynamicresource.CrudDomainService {
					return services.NewSalesVoucherDomainService(base)
				},
				NewAppServiceFn: app.NewSalesVoucherApplicationService,
			}, param)
		},
	)
}

// promotionReadOnlySchemas are served for reading only. Their domain service is the default and
// stays writable for the flows that own the rows (manage_* actions, Confirm, Cancel).
var promotionReadOnlySchemas = []string{
	models.SalesPromotionRuleProductSchemaName,
	models.SalesPromotionRuleCategorySchemaName,
	models.SalesPromotionRewardProductSchemaName,
	models.SalesPromotionRewardCategorySchemaName,
	models.SalesOrderVoucherSchemaName,
	models.SalesOrderPromotionSelectionSchemaName,
	models.SalesPromotionApplicationSchemaName,
	models.SalesPromotionAllocationSchemaName,
	models.SalesVoucherIssuanceSchemaName,
}

func registerPromotionReadOnlyEngines() error {
	errs := make([]error, 0, len(promotionReadOnlySchemas))
	for _, schemaName := range promotionReadOnlySchemas {
		name := schemaName
		errs = append(errs, deps.RegisterNamed(
			dynamicresource.EngineDependencyName(name),
			func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
				return buildOnion(&dynamicresource.DynamicResourceEngineOnionImpl{
					SchemaName:      name,
					NewRepositoryFn: childRepository,
					NewAppServiceFn: readOnlyAppService,
				}, param)
			},
		))
	}
	return stdErr.Join(errs...)
}

// forcePromotionEngines builds the promotion onions that publish no typed layer of their own.
func forcePromotionEngines() error {
	_, err := dynamicresource.ResolveEngines(PromotionEngineSchemaNames()[1:]...)
	return err
}

// PromotionEngineSchemaNames lists every promotion resource with an onion, in registration order.
func PromotionEngineSchemaNames() []string {
	return append([]string{
		models.SalesPromotionSchemaName, models.SalesPromotionRuleSchemaName, models.SalesPromotionRewardSchemaName,
		models.SalesVoucherSchemaName,
	}, promotionReadOnlySchemas...)
}
