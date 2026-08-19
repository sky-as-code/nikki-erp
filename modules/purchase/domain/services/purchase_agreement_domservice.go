package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The agreement lifecycle (BR §32-§46): confirm, close, cancel, and create_rfq.
//
// An agreement is a standing arrangement with a vendor — a blanket order committing to quantities
// at agreed prices, or a reusable template with no commitment attached. Orders are drawn against
// it; it is not itself an order.
//
// Unlike purchase_order, the agreement IS archivable. Archiving and restoring go through the
// engine's built-in set_archived rather than actions of their own: restore is the same power
// applied in reverse, so splitting them would let a role archive agreements it could not bring back.

// NewPurchaseAgreementDomainService derives the agreement service from the engine's default one.
func NewPurchaseAgreementDomainService(
	base drif.DynamicResourceService, references *OrderReferenceValidator,
) *PurchaseAgreementDomainServiceImpl {
	return &PurchaseAgreementDomainServiceImpl{DynamicResourceService: base, references: references}
}

type PurchaseAgreementDomainServiceImpl struct {
	drif.DynamicResourceService

	references *OrderReferenceValidator
}

var _ drif.DynamicResourceService = (*PurchaseAgreementDomainServiceImpl)(nil)

// Create stamps the fields a client may not choose, mirroring the order's rule.
//
// The code and the status are both no_update and both server-owned: an agreement that could be
// created `confirmed` would be a commitment nobody made, and one whose code a client picked could
// collide with or impersonate another.
func (this *PurchaseAgreementDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	code, err := generateAgreementCode()
	if err != nil {
		return nil, err
	}

	prepared := make(dmodel.DynamicFields, len(params)+3)
	for key, value := range params {
		prepared[key] = value
	}
	prepared[models.AgreementFieldCode] = code
	prepared[models.AgreementFieldStatus] = string(models.AgreementStatusDraft)

	// The vendor is OPTIONAL on an agreement where it is required on an order: a purchase template
	// is a reusable skeleton that may exist before anyone has chosen who to buy from. When one is
	// named it still has to be orderable.
	if this.references != nil && stringOf(prepared, models.AgreementFieldVendorId) != "" {
		vErrs := ft.NewClientErrors()
		if err := this.references.PrepareAgreement(ctx, prepared, vErrs); err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
		}
	}

	return this.DynamicResourceService.Create(ctx, prepared)
}

// Confirm makes the agreement live, so orders may be drawn against it (BR §38).
func (this *PurchaseAgreementDomainServiceImpl) Confirm(
	ctx corectx.Context, agreementId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.transitionAgreement(ctx, agreementId, agreementTransition{
		To:     string(models.AgreementStatusConfirmed),
		Action: AuditActionConfirm,
		Before: assertAgreementHasLines,
	})
}

// Close ends an agreement that has run its course (BR §42).
//
// The non-obvious part is the open-order guard. Closing an agreement with orders still open against
// it would strand them: they reference terms that are no longer in force, and nothing would say
// whether those terms still apply to the goods on their way. Cancel or confirm them first.
//
// The orders already CONFIRMED against it are untouched, and deliberately: closing says "no more
// orders from here", not "the ones already placed are void".
func (this *PurchaseAgreementDomainServiceImpl) Close(
	ctx corectx.Context, agreementId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.transitionAgreement(ctx, agreementId, agreementTransition{
		To:     string(models.AgreementStatusClosed),
		Action: AuditActionClose,
		Before: assertNoOpenOrders,
	})
}

// Cancel calls the agreement off (BR §43).
//
// It carries the same open-order guard as close, for the same reason: an order drawn against a
// cancelled agreement quotes terms that were withdrawn.
func (this *PurchaseAgreementDomainServiceImpl) Cancel(
	ctx corectx.Context, agreementId string, reason string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	return this.transitionAgreement(ctx, agreementId, agreementTransition{
		To:     string(models.AgreementStatusCancelled),
		Action: AuditActionCancel,
		Reason: reason,
		Before: assertNoOpenOrders,
	})
}

// agreementTransition describes one agreement status change.
type agreementTransition struct {
	To     string
	Action string
	Reason string

	// Before is an extra precondition, run inside the transaction once the record is loaded.
	Before func(ctx corectx.Context, agreement dmodel.DynamicFields) (*ft.ClientErrors, error)
}

func (this *PurchaseAgreementDomainServiceImpl) transitionAgreement(
	ctx corectx.Context, agreementId string, request agreementTransition,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		agreement, err := loadAgreement(tranxCtx, agreementId)
		if err != nil {
			return err
		}
		if agreement == nil {
			result = agreementNotFoundResult(agreementId)
			return nil
		}

		status := stringOf(agreement, models.AgreementFieldStatus)
		if status == request.To {
			// A retry after a lost response is not an error.
			result = mutateOk()
			return nil
		}

		vErrs := ft.NewClientErrors()
		AssertAgreementTransition(status, request.To, vErrs)
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
			return nil
		}

		if request.Before != nil {
			refusal, err := request.Before(tranxCtx, agreement)
			if err != nil {
				return err
			}
			if refusal != nil && refusal.Count() > 0 {
				result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *refusal}
				return nil
			}
		}

		if err := writeAgreementChanges(tranxCtx, agreement, dmodel.DynamicFields{
			models.AgreementFieldStatus: request.To,
		}); err != nil {
			return err
		}

		result = mutateOk()
		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.AgreementSchemaName,
			EntityId:   agreementId,
			Action:     request.Action,
			FromStatus: status,
			ToStatus:   request.To,
			Reason:     request.Reason,
			OrgId:      stringOf(agreement, basemodel.FieldOrgId),
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// assertAgreementHasLines refuses to confirm an agreement that commits to nothing.
//
// A blanket order with no line commits to no quantity at no price, which is not an agreement any
// vendor could act on. It is far more likely a half-filled form.
func assertAgreementHasLines(
	ctx corectx.Context, agreement dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	lineEngine, err := engineFor(models.AgreementLineSchemaName)
	if err != nil {
		return nil, err
	}
	agreementId := stringOf(agreement, models.AgreementFieldId)
	lines, err := models.FindAgreementLines(
		ctx, lineEngine.ResourceRepository(), agreementId, 1)
	if err != nil {
		return nil, err
	}
	if len(lines) > 0 {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.AgreementSchemaName,
		"purchase_agreement.no_lines",
		"a purchase agreement with no line cannot be confirmed"))
	return vErrs, nil
}

// assertNoOpenOrders refuses to close or cancel an agreement while orders are still open against
// it (BR §42).
func assertNoOpenOrders(
	ctx corectx.Context, agreement dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orderEngine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	agreementId := stringOf(agreement, models.AgreementFieldId)

	open, err := models.FindOpenOrdersForAgreement(
		ctx, orderEngine.ResourceRepository(), agreementId, 1)
	if err != nil {
		return nil, err
	}
	if len(open) == 0 {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.AgreementSchemaName,
		"purchase_agreement.has_open_orders",
		"this agreement still has orders open against it; confirm or cancel them first"))
	return vErrs, nil
}

// OrderedQuantities returns how much of each agreement line has been ordered (BR §41).
//
// It is DERIVED on read, never stored. The number changes whenever any referencing order is
// confirmed or cancelled, so a stored copy would need invalidating from the order side — and the
// day one of those paths forgot, the agreement would quietly under- or over-report its own
// drawdown with nothing to reconcile against.
//
// Only CONFIRMED orders count. A draft RFQ is a question, not a commitment, and counting it would
// show an agreement as drawn down by orders that may never be placed. A cancelled one is a
// commitment that was withdrawn.
//
// The result is keyed by product variant rather than by line: an agreement line commits to a
// quantity of a product, and an order drawn against the agreement names the product, not the line.
func OrderedQuantities(
	ctx corectx.Context, agreementId string,
) (map[string]OrderedQuantity, error) {
	orderEngine, err := engineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	lineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return nil, err
	}

	orders, err := models.FindConfirmedOrdersForAgreement(
		ctx, orderEngine.ResourceRepository(), agreementId, models.MaxAgreementOrders)
	if err != nil {
		return nil, err
	}

	ordered := map[string]OrderedQuantity{}
	for _, order := range orders {
		orderId := stringOf(order, models.PurchaseOrderFieldId)
		lines, err := models.FindOrderLines(
			ctx, lineEngine.ResourceRepository(), orderId, models.MaxOrderLines)
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			variantId := stringOf(line, models.PurchaseOrderLineFieldProductVariantId)
			if variantId == "" {
				continue
			}
			// The ORDERED quantity in the line's own unit, not the inventory one: the agreement
			// commits to a quantity in a stated unit, and comparing against a converted number
			// would compare two different things.
			ordered[variantId] = OrderedQuantity{
				Quantity: ordered[variantId].Quantity.Add(
					decimalOf(line, models.PurchaseOrderLineFieldQuantity)),
				UomId: stringOf(line, models.PurchaseOrderLineFieldUomId),
			}
		}
	}
	return ordered, nil
}

// OrderedQuantity is an amount with the unit it is expressed in. The unit travels with the number
// because a quantity without one cannot be compared to anything.
type OrderedQuantity struct {
	Quantity decimal.Decimal
	UomId    string
}

func loadAgreement(ctx corectx.Context, agreementId string) (dmodel.DynamicFields, error) {
	engine, err := engineFor(models.AgreementSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.AgreementFieldId: agreementId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadAgreement")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

func writeAgreementChanges(
	ctx corectx.Context, agreement dmodel.DynamicFields, changes dmodel.DynamicFields,
) error {
	engine, err := engineFor(models.AgreementSchemaName)
	if err != nil {
		return err
	}

	update := make(dmodel.DynamicFields, len(changes)+2)
	for key, value := range changes {
		update[key] = value
	}
	update[models.AgreementFieldId] = stringOf(agreement, models.AgreementFieldId)
	update[basemodel.FieldEtag] = stringOf(agreement, basemodel.FieldEtag)

	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeAgreementChanges")
}

func agreementNotFoundResult(agreementId string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.AgreementSchemaName,
		"purchase_agreement.not_found",
		"no purchase agreement with id '"+agreementId+"'"))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// generateAgreementCode mints the agreement's human-facing reference, on the order code's pattern.
func generateAgreementCode() (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "generateAgreementCode")
	}
	return "PA-" + *id, nil
}

// CreateRfq starts a new request for quotation from an agreement (BR §40).
//
// The agreement's lines become the order's lines, at the agreed quantities and prices — that is the
// point of a blanket order, and of a template. What the new order does NOT inherit is the
// agreement's status or history: it starts as an ordinary draft RFQ that happens to have been
// pre-filled, and it must be confirmed like any other.
//
// Only a CONFIRMED agreement may raise one. A draft has not been agreed with anyone, so an order
// quoting its prices would quote terms nobody committed to; a closed or cancelled one has been
// withdrawn.
func (this *PurchaseAgreementDomainServiceImpl) CreateRfq(
	ctx corectx.Context, agreementId string, orderService *PurchaseOrderDomainServiceImpl,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	var result *dyn.OpResult[dmodel.DynamicFields]

	err := withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		agreement, err := loadAgreement(tranxCtx, agreementId)
		if err != nil {
			return err
		}
		if agreement == nil {
			notFound := agreementNotFoundResult(agreementId)
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: notFound.ClientErrors}
			return nil
		}

		status := stringOf(agreement, models.AgreementFieldStatus)
		if status != string(models.AgreementStatusConfirmed) {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation(models.AgreementSchemaName,
				"purchase_agreement.not_confirmed",
				"only a confirmed purchase agreement can raise a request for quotation; "+
					"this one is '"+status+"'"))
			result = &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}
			return nil
		}

		created, err := orderService.Create(tranxCtx, orderFromAgreement(agreement, agreementId))
		if err != nil || created.ClientErrors.Count() > 0 {
			result = created
			return err
		}
		result = created

		orderId := stringOf(created.Data, models.PurchaseOrderFieldId)
		if err := this.copyAgreementLines(tranxCtx, agreementId, orderId,
			stringOf(agreement, basemodel.FieldOrgId)); err != nil {
			return err
		}
		if err := RecomputeOrderTotals(tranxCtx, orderId); err != nil {
			return err
		}

		return WriteAuditEvent(tranxCtx, AuditEntry{
			EntityType: models.PurchaseOrderSchemaName,
			EntityId:   orderId,
			Action:     AuditActionCreateRfq,
			ToStatus:   string(models.PurchaseOrderStatusRfq),
			OrgId:      stringOf(agreement, basemodel.FieldOrgId),
			Metadata:   map[string]any{"from_agreement": agreementId},
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// orderFromAgreement carries the agreement's terms onto the new order.
//
// agreement_id is set, which is what makes the order countable against the agreement's drawdown
// (BR §41) and what the close guard looks for.
func orderFromAgreement(agreement dmodel.DynamicFields, agreementId string) dmodel.DynamicFields {
	params := dmodel.DynamicFields{
		models.PurchaseOrderFieldAgreementId: agreementId,
	}
	for _, pair := range []struct{ from, to string }{
		{models.AgreementFieldVendorId, models.PurchaseOrderFieldVendorId},
		{models.AgreementFieldBuyerId, models.PurchaseOrderFieldBuyerId},
		{models.AgreementFieldCurrencyId, models.PurchaseOrderFieldCurrencyId},
		{basemodel.FieldOrgId, basemodel.FieldOrgId},
	} {
		if value, ok := agreement[pair.from]; ok && value != nil {
			params[pair.to] = value
		}
	}
	return params
}

// copyAgreementLines turns each agreement line into an order line at the agreed price.
func (this *PurchaseAgreementDomainServiceImpl) copyAgreementLines(
	ctx corectx.Context, agreementId, orderId, orgId string,
) error {
	agreementLineEngine, err := engineFor(models.AgreementLineSchemaName)
	if err != nil {
		return err
	}
	orderLineEngine, err := engineFor(models.PurchaseOrderLineSchemaName)
	if err != nil {
		return err
	}

	lines, err := models.FindAgreementLines(
		ctx, agreementLineEngine.ResourceRepository(), agreementId, models.MaxAgreementLines)
	if err != nil {
		return err
	}

	for _, line := range lines {
		orderLine := dmodel.DynamicFields{
			models.PurchaseOrderLineFieldPurchaseOrderId:  orderId,
			models.PurchaseOrderLineFieldLineType:         string(models.PurchaseOrderLineTypeProduct),
			models.PurchaseOrderLineFieldSequence:         line[models.AgreementLineFieldSequence],
			models.PurchaseOrderLineFieldProductVariantId: line[models.AgreementLineFieldProductVariantId],
			models.PurchaseOrderLineFieldUomId:            line[models.AgreementLineFieldUomId],
			models.PurchaseOrderLineFieldQuantity:         line[models.AgreementLineFieldQuantity],
			models.PurchaseOrderLineFieldUnitPrice:        line[models.AgreementLineFieldUnitPrice],
			models.PurchaseOrderLineFieldDescription:      line[models.AgreementLineFieldDescription],
			// An agreement carries no tax or discount, so the new line starts with neither. Tax is
			// an input the buyer supplies per order (D9), not a term of the agreement.
			models.PurchaseOrderLineFieldDiscountPercent: decimal.Zero,
			models.PurchaseOrderLineFieldTaxAmount:       decimal.Zero,
		}
		if orgId != "" {
			orderLine[basemodel.FieldOrgId] = orgId
		}

		if _, err := orderLineEngine.ResourceService().Create(ctx, orderLine); err != nil {
			return err
		}
	}
	return nil
}
