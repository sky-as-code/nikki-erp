// Package dynamicengines declares the resource engines the Sales module serves, and creates them
// during Init(). It may import only the domain, the module's own interfaces and dynamicresource —
// never app/, infra/ or transport/ — so both sales and sales/transport/restful can import it without
// a cycle. The rules its callbacks enforce live in domain/services.
package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// engineSpec declares one resource engine the Sales module owns.
type engineSpec struct {
	// SchemaName must be an XSchemaName constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions is optional; a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources this module serves. The order matches RegisterModels for
// readability only; engines are created after every schema is registered. Junction tables never get
// an entry: a _rel row is configured through its owner, so it has no route and no IAM resource row.
var engineSpecs = []engineSpec{
	salesChannelEngineSpec(),
	salesPointEngineSpec(),
	salesOrderEngineSpec(),
	salesOrderLineEngineSpec(),
	salesOrderLineComponentEngineSpec(),
	salesOrderAdjustmentEngineSpec(),
	salesOrderEventEngineSpec(),

	// Pricing master data, client-managed: an operator sets up pricelists and bundles through the UI.
	salesPricelistEngineSpec(),
	salesPricelistItemEngineSpec(),
	salesComboEngineSpec(),
	salesComboComponentEngineSpec(),

	// Promotion master data, also operator-managed.
	salesPromotionProgramEngineSpec(),
	salesPromotionConditionGroupEngineSpec(),
	salesPromotionConditionEngineSpec(),
	salesPromotionConditionTargetEngineSpec(),
	salesPromotionRewardEngineSpec(),
	salesPromotionCompatibilityEngineSpec(),

	// Voucher codes are operator-managed master data; redemptions are the ledger of their use.
	salesVoucherCodeEngineSpec(),
	salesVoucherRedemptionEngineSpec(),

	// Billing: the settlement units of a sale, their allocations, and the lineage between them.
	salesBillEngineSpec(),
	salesBillLineEngineSpec(),
	salesBillRelationEngineSpec(),
	salesPaymentEngineSpec(),
	salesFulfillmentRequestEngineSpec(),
	salesFulfillmentRequestLineEngineSpec(),

	// The fiscal contract: what Sales asked an eInvoice provider for, and what came back.
	salesFiscalRequestEngineSpec(),

	// Billing instructions: who a sale is to be invoiced to, and the record of each try at issuing.
	salesBillingInstructionEngineSpec(),
	salesBillingIssuanceAttemptEngineSpec(),

	// Operator price overrides.
	salesManualDiscountEngineSpec(),

	// Quotations: an offer and its lines, which back-office raises and converts.
	salesQuotationEngineSpec(),
	salesQuotationLineEngineSpec(),

	// Returns and the refunds that settle them.
	salesReturnEngineSpec(),
	salesReturnLineEngineSpec(),
	salesRefundPaymentEngineSpec(),

	// The integration event feed.
	salesIntegrationOutboxEngineSpec(),
}

// junctionSchemas lists association schemas that need an engine built but not served: a schema here
// gets a repository and nothing else — no route, no IAM resource row, no CRUD. They need an engine
// only because a repository cannot be built without one (the query builder and database client are
// private to the registry). crud.ManageM2m is not usable here because it resolves the far side
// against a locally registered schema, and sales_channel_payment_rel points at a paymentinvoice
// payment method.
var junctionSchemas = []string{
	models.SalesChannelPaymentRelSchemaName,
}

func salesChannelEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesChannelSchemaName,
		DefineActions: func(engine drif.DynamicResourceEngine) error {
			// Two families of action on the same record: lifecycle and payment-method configuration.
			// Declared separately because they answer to different permissions and seed rows.
			return stdErr.Join(
				defineSalesChannelActions(engine),
				defineChannelPaymentActions(engine),
			)
		},
	}
}

// salesOrderEngineSpec serves the order. The lifecycle actions wait for the state machines that
// would validate them; apply_voucher is safe ahead of those because it moves no status, and party
// assignment because it names who is party to a sale rather than moving the sale itself.
func salesOrderEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderSchemaName,
		DefineActions: func(engine drif.DynamicResourceEngine) error {
			// Two families of action on the same record: pricing and the parties to the sale.
			// Declared separately because they answer to different permissions and seed rows.
			return stdErr.Join(
				defineSalesOrderVoucherActions(engine),
				defineSalesOrderPartyActions(engine),
			)
		},
	}
}

// A line is its own resource rather than a nested payload, because a whole-order update would
// rewrite untouched lines and lose concurrent edits to them.
func salesOrderLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderLineSchemaName,
	}
}

// The three records that explain an order rather than form it, all read-only: a client able to POST
// one could forge a price explanation or an audit trail. engineSpec has no ReadOnly flag, so the
// read-onlyness comes from their IAM seeds granting `read` alone.
func salesOrderLineComponentEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderLineComponentSchemaName,
	}
}

func salesOrderAdjustmentEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderAdjustmentSchemaName,
	}
}

func salesOrderEventEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderEventSchemaName,
	}
}

func salesPricelistEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesPricelistSchemaName,
		DefineActions: defineSalesPricelistActions,
	}
}

func salesPricelistItemEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPricelistItemSchemaName}
}

func salesComboEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesComboSchemaName}
}

func salesComboComponentEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesComboComponentSchemaName}
}

func salesPromotionProgramEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionProgramSchemaName}
}

func salesPromotionConditionGroupEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionConditionGroupSchemaName}
}

func salesPromotionConditionEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionConditionSchemaName}
}

func salesPromotionConditionTargetEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionConditionTargetSchemaName}
}

func salesPromotionRewardEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionRewardSchemaName}
}

func salesBillEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesBillSchemaName,
		DefineActions: defineSalesBillActions,
	}
}

func salesBillLineEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesBillLineSchemaName}
}

// The lineage is read-only: its rows come from split and merge alone, and a writable one could
// fabricate a trail showing a payment settled a bill it never touched.
func salesBillRelationEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesBillRelationSchemaName}
}

// Payments are read-only: money is recorded through record_payment alone, which applies gates a
// plain POST would bypass, including asking another module whether the method may be used.
func salesPaymentEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPaymentSchemaName}
}

// The fulfilment tables are read-only: a writable one could tell Inventory to move goods no sale
// asked for.
func salesFulfillmentRequestEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesFulfillmentRequestSchemaName}
}

func salesFulfillmentRequestLineEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesFulfillmentRequestLineSchemaName}
}

// The fiscal contract is read-only: a writable one could ask a tax authority for a document against
// a sale that never happened, or mark an unissued request as issued.
func salesFiscalRequestEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesFiscalRequestSchemaName,
		DefineActions: defineSalesFiscalRequestActions,
	}
}

// A billing instruction is client-creatable: recording that a buyer wants an invoice is how the
// process starts, at a till or through back office. Its status is no_update, so the lifecycle runs
// only through the actions below — a client that could write `ready` directly would be able to
// release a document for issuance without the completeness check.
func salesBillingInstructionEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesBillingInstructionSchemaName,
		DefineActions: defineSalesBillingInstructionActions,
	}
}

// Issuance attempts are read-only: they are the evidence of whether a document was created,
// including the indeterminate case where nobody knows. A writable attempt could fabricate a record
// showing an invoice was issued, or erase the trace of one that may exist.
func salesBillingIssuanceAttemptEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesBillingIssuanceAttemptSchemaName}
}

// A return is client-creatable, unlike a bill: raising one is how an agent starts the process. Its
// status columns are no_update, so the lifecycle runs only through the actions below.
func salesReturnEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesReturnSchemaName,
		DefineActions: defineSalesReturnActions,
	}
}

// Return lines are read-only: create_return checks each quantity against what is still returnable
// and prices it from historical amounts, so a writable line could return more than was delivered or
// name its own refund amount.
func salesReturnLineEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesReturnLineSchemaName}
}

// Refund legs are read-only: the return workflow caps each leg at what its original payment
// captured, so a writable row could create an outflow with no matching inflow.
func salesRefundPaymentEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesRefundPaymentSchemaName}
}

// The outbox is read-only: a writable row could announce a sale that never happened, or set
// published_at on an event that never went, which drops it from the queue. It is routed at all so an
// operator can see what Sales believes it published, and when.
func salesIntegrationOutboxEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesIntegrationOutboxSchemaName}
}

// Overrides are read-only: a plain POST would bypass the gates that make one auditable (mandatory
// reason, draft-only check, audit entry recording both prices).
func salesManualDiscountEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesManualDiscountSchemaName}
}

// A quotation is client-creatable through the built-in POST, since creating one commits nothing.
// Its status is no_update and moves only through the actions below, because accepting one creates
// an order.
func salesQuotationEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesQuotationSchemaName,
		DefineActions: defineSalesQuotationActions,
	}
}

func salesQuotationLineEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesQuotationLineSchemaName}
}

func salesVoucherCodeEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesVoucherCodeSchemaName}
}

// The redemption ledger is read-only through its IAM seed granting `read` alone: a writable row
// could forge a discount's provenance or release a hold another order relies on.
func salesVoucherRedemptionEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesVoucherRedemptionSchemaName}
}

func salesPromotionCompatibilityEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPromotionCompatibilitySchemaName}
}

func salesPointEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesPointSchemaName,
		DefineActions: defineSalesPointActions,
	}
}

// EngineSchemaNames keeps route registration and engine creation from drifting apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates this module's engines and publishes them into the dependency container
// so other modules can inject them by name.
func InitDynamicEngines() error {
	for _, spec := range engineSpecs {
		if err := initEngine(spec); err != nil {
			return err
		}
	}
	for _, schemaName := range junctionSchemas {
		if err := initEngine(engineSpec{SchemaName: schemaName}); err != nil {
			return err
		}
	}
	return nil
}

func initEngine(spec engineSpec) error {
	engine, err := dynamicresource.Registry().NewEngine(spec.SchemaName, drif.NewEngineOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create the '%s' resource engine", spec.SchemaName)
	}

	if spec.DefineActions != nil {
		if err := spec.DefineActions(engine); err != nil {
			return errors.Wrapf(err, "failed to define actions of the '%s' resource engine", spec.SchemaName)
		}
	}

	err = deps.RegisterNamed(
		dynamicresource.EngineDependencyName(spec.SchemaName),
		func() drif.DynamicResourceEngine { return engine },
	)
	return errors.Wrapf(err, "failed to register the '%s' resource engine", spec.SchemaName)
}
