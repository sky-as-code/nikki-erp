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
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itFiscal "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fiscal"
)

// The fiscal resources: the request for an invoice, the instruction that says how to bill, and the
// read-only record of each issuance attempt.
//
// The invoicing port has no adapter in every deployment. It is injected as an ordinary dependency
// because the container registers a nil-tolerant implementation; the operation checks for it and
// answers a refusal rather than failing, which is why a deployment without e-invoicing still
// takes orders.

type salesFiscalRequestEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_fiscal_request"`
}

func registerSalesFiscalRequestEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesFiscalRequestSchemaName),
		func(
			param composable.BuildParam,
			invoicing itInvoicing.InvoicingExtService,
			settings itExt.EffectiveSettingsExtService,
			parties itExt.PartyExtService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesFiscalRequestSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesFiscalRequestRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesFiscalRequestApplicationService(base, invoicing, settings, parties)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesFiscalRequestEngineParam) itFiscal.SalesFiscalRequestRepository {
			return p.Engine.Repository().(itFiscal.SalesFiscalRequestRepository)
		},
		func(p salesFiscalRequestEngineParam) itFiscal.SalesFiscalRequestApplicationService {
			return p.Engine.ApplicationService().(itFiscal.SalesFiscalRequestApplicationService)
		},
	))
}

type salesBillingInstructionEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_billing_instruction"`
}

func registerSalesBillingInstructionEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesBillingInstructionSchemaName),
		func(
			param composable.BuildParam,
			parties itExt.PartyExtService,
			settings itExt.EffectiveSettingsExtService,
		) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName: models.SalesBillingInstructionSchemaName,
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesBillingInstructionRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesBillingInstructionApplicationService(base, parties, settings)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesBillingInstructionEngineParam) itFiscal.SalesBillingInstructionRepository {
			return p.Engine.Repository().(itFiscal.SalesBillingInstructionRepository)
		},
		func(p salesBillingInstructionEngineParam) itFiscal.SalesBillingInstructionApplicationService {
			return p.Engine.ApplicationService().(itFiscal.SalesBillingInstructionApplicationService)
		},
	))
}

type salesBillingIssuanceAttemptEngineParam struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_sales_billing_issuance_attempt"`
}

// The issuance attempt records what a provider answered, so it is read-only.
func registerSalesBillingIssuanceAttemptEngine() error {
	err := deps.RegisterNamed(
		composable.EngineDependencyName(models.SalesBillingIssuanceAttemptSchemaName),
		func(param composable.BuildParam) composable.DynamicResourceEngineOnion {
			return buildOnion(&composable.DynamicResourceEngineOnionImpl{
				SchemaName:  models.SalesBillingIssuanceAttemptSchemaName,
				CrudActions: readOnlyCrudActions(),
				NewRepositoryFn: func(base composable.CrudRepository) composable.CrudRepository {
					return repo.NewSalesBillingIssuanceAttemptRepository(base)
				},
				NewAppServiceFn: func(base composable.CrudApplicationService) composable.CrudApplicationService {
					return app.NewSalesBillingIssuanceAttemptApplicationService(base)
				},
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p salesBillingIssuanceAttemptEngineParam) itFiscal.SalesBillingIssuanceAttemptRepository {
			return p.Engine.Repository().(itFiscal.SalesBillingIssuanceAttemptRepository)
		},
		func(p salesBillingIssuanceAttemptEngineParam) itFiscal.SalesBillingIssuanceAttemptApplicationService {
			return p.Engine.ApplicationService().(itFiscal.SalesBillingIssuanceAttemptApplicationService)
		},
	))
}
