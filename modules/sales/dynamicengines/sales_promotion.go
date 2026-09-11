package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	repo "github.com/sky-as-code/nikki-erp/modules/sales/infra/repository"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/promotion"
)

// The promotion and voucher resources: plain CRUD master data the pricing engine reads.

type salesPromotionProgramEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_program"`
}

func registerSalesPromotionProgramEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionProgramSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionProgramSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionProgramRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionProgramApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionProgramEngineParam) itPromotion.SalesPromotionProgramRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionProgramRepository)
		},
		func(p salesPromotionProgramEngineParam) itPromotion.SalesPromotionProgramApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionProgramApplicationService)
		},
	))
}

type salesPromotionConditionGroupEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_group"`
}

func registerSalesPromotionConditionGroupEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionConditionGroupSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionConditionGroupSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionConditionGroupRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionConditionGroupApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionConditionGroupEngineParam) itPromotion.SalesPromotionConditionGroupRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionConditionGroupRepository)
		},
		func(p salesPromotionConditionGroupEngineParam) itPromotion.SalesPromotionConditionGroupApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionConditionGroupApplicationService)
		},
	))
}

type salesPromotionConditionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition"`
}

func registerSalesPromotionConditionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionConditionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionConditionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionConditionRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionConditionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionConditionEngineParam) itPromotion.SalesPromotionConditionRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionConditionRepository)
		},
		func(p salesPromotionConditionEngineParam) itPromotion.SalesPromotionConditionApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionConditionApplicationService)
		},
	))
}

type salesPromotionConditionTargetEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_condition_target"`
}

func registerSalesPromotionConditionTargetEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionConditionTargetSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionConditionTargetSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionConditionTargetRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionConditionTargetApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionConditionTargetEngineParam) itPromotion.SalesPromotionConditionTargetRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionConditionTargetRepository)
		},
		func(p salesPromotionConditionTargetEngineParam) itPromotion.SalesPromotionConditionTargetApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionConditionTargetApplicationService)
		},
	))
}

type salesPromotionRewardEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_reward"`
}

func registerSalesPromotionRewardEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionRewardSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionRewardSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionRewardRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionRewardApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionRewardEngineParam) itPromotion.SalesPromotionRewardRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionRewardRepository)
		},
		func(p salesPromotionRewardEngineParam) itPromotion.SalesPromotionRewardApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionRewardApplicationService)
		},
	))
}

type salesPromotionCompatibilityEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_promotion_compatibility"`
}

func registerSalesPromotionCompatibilityEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesPromotionCompatibilitySchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesPromotionCompatibilitySchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesPromotionCompatibilityRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesPromotionCompatibilityApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesPromotionCompatibilityEngineParam) itPromotion.SalesPromotionCompatibilityRepository {
			return p.Engine.Repository().(itPromotion.SalesPromotionCompatibilityRepository)
		},
		func(p salesPromotionCompatibilityEngineParam) itPromotion.SalesPromotionCompatibilityApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesPromotionCompatibilityApplicationService)
		},
	))
}

type salesVoucherCodeEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_code"`
}

func registerSalesVoucherCodeEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesVoucherCodeSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesVoucherCodeSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesVoucherCodeRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesVoucherCodeApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesVoucherCodeEngineParam) itPromotion.SalesVoucherCodeRepository {
			return p.Engine.Repository().(itPromotion.SalesVoucherCodeRepository)
		},
		func(p salesVoucherCodeEngineParam) itPromotion.SalesVoucherCodeApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesVoucherCodeApplicationService)
		},
	))
}

// The redemption ledger is read-only: a writable row could forge a discount's provenance or
// release a hold another order relies on.
type salesVoucherRedemptionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_voucher_redemption"`
}

func registerSalesVoucherRedemptionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesVoucherRedemptionSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesVoucherRedemptionSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesVoucherRedemptionRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesVoucherRedemptionApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesVoucherRedemptionEngineParam) itPromotion.SalesVoucherRedemptionRepository {
			return p.Engine.Repository().(itPromotion.SalesVoucherRedemptionRepository)
		},
		func(p salesVoucherRedemptionEngineParam) itPromotion.SalesVoucherRedemptionApplicationService {
			return p.Engine.ApplicationService().(itPromotion.SalesVoucherRedemptionApplicationService)
		},
	))
}
