// Package dynamicengines declares the resource engines the Sales module serves through the
// dynamic resource engine, and creates them during the module's Init().
//
// Its imports point one way only: the domain, the module's own interfaces, and the dynamicresource
// module — never app/, infra/ or transport/. That keeps the package importable by both sales
// (which creates the engines) and sales/transport/restful (which registers their routes) without
// a cycle.
//
// The package declares engines and adapts their callbacks; the rules those callbacks enforce live
// in domain/services. See docs/wiki/07. ERP backend module.md §6.7.
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
	// SchemaName is the dynamic-model schema the engine serves. It must be an XSchemaName
	// constant, never a string derived from the resource path.
	SchemaName string

	// DefineActions adds resource-specific actions and validation on top of the built-in CRUD
	// ones. It is optional: a resource without custom behavior leaves it nil.
	DefineActions func(drif.DynamicResourceEngine) error
}

// engineSpecs lists the resources this module serves through the dynamic resource engine.
//
// The order matches RegisterModels: referenced before referencing. It does not have to — engines
// are created after every schema is registered — but keeping the two lists in the same order makes
// a missing entry obvious when reading them side by side.
//
// Junction tables never get an entry here. A _rel row is configured through its owner's
// capabilities, so it has no route and no IAM resource row — see the 25-engines-for-27-schemas
// split in vending_machine_new.
var engineSpecs = []engineSpec{
	salesChannelEngineSpec(),
	salesPointEngineSpec(),
	salesOrderEngineSpec(),
	salesOrderLineEngineSpec(),
	salesOrderLineComponentEngineSpec(),
	salesOrderAdjustmentEngineSpec(),
	salesOrderEventEngineSpec(),

	// Pricing master data. These ARE client-managed, unlike the three explanatory records above:
	// an operator sets up pricelists and bundles through the UI, which is the whole point of
	// having them as data rather than as code.
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

	// Operator price overrides.
	salesManualDiscountEngineSpec(),

	// Quotations: an offer and its lines, which back-office raises and converts.
	salesQuotationEngineSpec(),
	salesQuotationLineEngineSpec(),

	// The integration event feed.
	salesIntegrationOutboxEngineSpec(),
}

// junctionSchemas lists the association schemas that need an engine built but not served.
//
// The distinction from engineSpecs is the whole point: EngineSchemaNames drives route registration,
// so a schema here gets a repository and nothing else — no route, no IAM resource row, no CRUD.
//
// They need an engine at all only because a repository cannot be built without one: the query
// builder and database client the registry injects are private to it. vending_machine_new avoids
// this for its own junctions by writing them through crud.ManageM2m on the owner's repository, but
// that helper resolves the far side against a locally registered schema. sales_channel_payment_rel
// points at a paymentinvoice payment method, which is not one, so its rows are written directly and
// a repository of its own is what makes that possible.
var junctionSchemas = []string{
	models.SalesChannelPaymentRelSchemaName,
}

func salesChannelEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesChannelSchemaName,
		DefineActions: func(engine drif.DynamicResourceEngine) error {
			// The channel carries two families of action: its own lifecycle, and the payment-method
			// configuration that belongs to it rather than to a resource of its own. They are
			// declared separately because they answer to different permissions and are seeded as
			// different rows, but they hang off the same engine because they act on the same record.
			return stdErr.Join(
				defineSalesChannelActions(engine),
				defineChannelPaymentActions(engine),
			)
		},
	}
}

// salesOrderEngineSpec serves the order over HTTP, with apply_voucher as its only custom action.
//
// The lifecycle actions - confirm, cancel - arrive with SALES-013 and SALES-014, which is also when
// the state machines that make them safe exist. Declaring those routes now would expose transitions
// nothing yet validates.
//
// apply_voucher is safe ahead of them because it does not move any status: it reserves a use and
// reports what the basket should now be priced against, and a draft order is editable by definition.
func salesOrderEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesOrderSchemaName,
		DefineActions: defineSalesOrderVoucherActions,
	}
}

// salesOrderLineEngineSpec serves the lines.
//
// A line is a resource in its own right rather than only a nested payload of its order, because
// adding, changing and removing one line of a draft is the operation a POS screen performs on every
// keystroke; making that a whole-order update would rewrite untouched lines and lose a concurrent
// edit to them.
func salesOrderLineEngineSpec() engineSpec {
	return engineSpec{
		SchemaName: models.SalesOrderLineSchemaName,
	}
}

// The three records that explain an order rather than form it. All read-only over HTTP: each is
// written by the operation that produces it - combo expansion, the pricing engine, the audit
// service - and a client able to POST one could forge a price explanation or an audit trail.
//
// ReadOnly is not available on this module's engineSpec, so the write actions are simply not
// exposed: their IAM seeds grant read alone, which is what the engine asserts against.
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

// salesBillRelationEngineSpec routes the lineage read-only.
//
// Its rows are written by split and merge and by nothing else. A client able to POST one could
// fabricate a paper trail showing a payment settled a bill it never touched.
func salesBillRelationEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesBillRelationSchemaName}
}

// salesPaymentEngineSpec routes payments read-only.
//
// Money is recorded through the record_payment operation and by nothing else: it applies six gates
// that a plain POST would bypass, including the two that ask another module whether the method may
// be used at all.
func salesPaymentEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesPaymentSchemaName}
}

// The fulfilment tables are read-only over HTTP. Requests are raised by confirm and by the return
// workflow, and a client able to write one could tell Inventory to move goods no sale asked for.
func salesFulfillmentRequestEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesFulfillmentRequestSchemaName}
}

func salesFulfillmentRequestLineEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesFulfillmentRequestLineSchemaName}
}

// salesFiscalRequestEngineSpec routes the fiscal contract read-only.
//
// Requests are raised by the request_invoice operation and by nothing else. A client able to POST
// one could ask a tax authority for a legal document against a sale that was never made; a client
// able to PATCH one could mark an unissued request as issued, which is the state BR 77 exists to
// keep honest.
func salesFiscalRequestEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.SalesFiscalRequestSchemaName,
		DefineActions: defineSalesFiscalRequestActions,
	}
}

// salesIntegrationOutboxEngineSpec routes the outbox read-only.
//
// Rows are written by the domain services that produce the events, inside their own transactions,
// and drained by the sweep in app/. A client able to POST one could announce a sale that never
// happened to every downstream consumer; a client able to PATCH one could set published_at on an
// event that never went, which deletes it from the queue as surely as a DELETE would.
//
// It is routed at all because an operator investigating a consumer that is behind needs to see what
// Sales believes it published, and when.
func salesIntegrationOutboxEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesIntegrationOutboxSchemaName}
}

// salesManualDiscountEngineSpec routes the overrides read-only.
//
// Rows are written by the grant operation and by nothing else, because a plain POST would bypass
// every gate that makes an override auditable: the mandatory reason, the draft-only check, and the
// audit entry recording both prices. A client able to write one directly could change what a
// customer pays with no stated cause and no trail.
func salesManualDiscountEngineSpec() engineSpec {
	return engineSpec{SchemaName: models.SalesManualDiscountSchemaName}
}

// salesQuotationEngineSpec carries the convert and transition actions.
//
// Unlike the order, a quotation IS client-creatable through the built-in POST: it is a document an
// operator writes, and nothing about creating one commits the business or moves money. What is gated
// is the status - declared no_update, moved only through the actions below - because accepting one
// creates an order and expiring one closes an offer.
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

// salesVoucherRedemptionEngineSpec routes the redemption ledger read-only.
//
// It is registered like any other engine; its read-onlyness comes from the IAM seed granting `read`
// alone. A client able to POST one could forge a discount's provenance, and a client able to PATCH
// one could release a hold another order is relying on.
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

// EngineSchemaNames lists the schemas this module creates an engine for, so that route
// registration and engine creation cannot drift apart.
func EngineSchemaNames() []string {
	return array.Map(engineSpecs, func(spec engineSpec) string {
		return spec.SchemaName
	})
}

// InitDynamicEngines creates the resource engines this module owns and publishes them into the
// dependency container, so that other modules can inject them by name.
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
