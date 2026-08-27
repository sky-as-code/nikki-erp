package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Manual discount and price override (BR 87.4, SALES-039).
//
// # The base price is never overwritten
//
// The override is stored as its own row and replayed through the pricing engine as an adjustment on
// top. Overwriting base_unit_price would be simpler and is wrong twice: BR 87.9's explanation would
// show a chain starting from a price the catalogue never charged, and an override would become
// indistinguishable from a genuinely cheap product — so nobody could audit who had discounted what.
//
// # Why a stored row rather than a written adjustment
//
// Repricing DELETES the whole adjustment chain and rewrites it from engine output, and confirm
// reprices unconditionally. An adjustment written straight into sales_order_adjustments would
// therefore be erased before the sale completed, silently, with the customer charged full price
// after being promised a discount. Storing the override and feeding it back in on every calculation
// is what makes it survive.
//
// # Reason is mandatory, and the permission is checked in app/
//
// BR 87.4 gates this on permission. That check belongs in app/ like every other authorization, so
// this service records WHO granted it rather than deciding whether they could — but it does enforce
// the reason, because a reason is a business invariant rather than an access decision, and an
// override with no stated cause is indistinguishable from a mistake.

// GrantManualDiscountParams is what an override needs.
type GrantManualDiscountParams struct {
	SalesOrderId string

	// SalesOrderLineId is empty for an order-level override, which the engine spreads across the
	// lines proportionally.
	SalesOrderLineId string

	// Amount is POSITIVE — what comes off. A negative value is a surcharge, which BR 87.4 does not
	// authorise.
	Amount decimal.Decimal

	// Reason is mandatory. See the package comment.
	Reason string

	// GrantedBy is deliberately ABSENT. The actor comes from the context, like the audit trail's,
	// so that a caller cannot record somebody else as having authorised a discount - which is
	// exactly the claim an auditor is relying on when they read this field.
}

// GrantManualDiscountResult is what granting produced.
type GrantManualDiscountResult struct {
	SalesManualDiscountId string
	SalesOrderId          string

	// TotalBefore and TotalAfter are both returned so the caller can see what the override actually
	// did. The engine caps a discount at what is owed, so the difference is not always the amount
	// asked for — and an operator told only "granted" would not know that.
	TotalBefore decimal.Decimal
	TotalAfter  decimal.Decimal
}

// The refusal reasons granting an override can produce.
const (
	ReasonDiscountReasonRequired = "sales_manual_discount.reason_required"
	ReasonDiscountNotPositive    = "sales_manual_discount.amount_not_positive"
	ReasonOrderNotDiscountable   = "sales_manual_discount.order_not_discountable"
	ReasonDiscountLineNotFound   = "sales_manual_discount.line_not_found"
)

// GrantManualDiscount records an operator override and reprices the order (BR 87.4).
func GrantManualDiscount(
	ctx corectx.Context,
	params GrantManualDiscountParams,
	taxSvc itExt.TaxCalculationExtService,
	policy SalesPolicy,
) (*GrantManualDiscountResult, *ft.ClientErrors, error) {
	order, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return nil, nil, err
	}
	if order == nil {
		return nil, OrderNotFoundErrors(params.SalesOrderId), nil
	}

	vErrs, err := assertDiscountable(ctx, order, params)
	if err != nil {
		return nil, nil, err
	}
	if vErrs != nil {
		return nil, vErrs, nil
	}

	totalBefore := decimalOf(order, models.SalesOrderFieldGrandTotal)

	discountId, err := writeManualDiscount(ctx, order, params, totalBefore)
	if err != nil {
		return nil, nil, err
	}

	// Reprice immediately rather than leaving the order stale. The override only takes effect
	// through the engine, so an order that was not repriced would show the old total while carrying
	// a discount somebody had already authorised — and a till reading that total would charge it.
	if _, vErrs, err := RepriceOrder(ctx, params.SalesOrderId, taxSvc, policy); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	repriced, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, params.SalesOrderId)
	if err != nil {
		return nil, nil, err
	}

	return &GrantManualDiscountResult{
		SalesManualDiscountId: discountId,
		SalesOrderId:          params.SalesOrderId,
		TotalBefore:           totalBefore,
		TotalAfter:            decimalOf(repriced, models.SalesOrderFieldGrandTotal),
	}, nil, nil
}

// assertDiscountable applies the gates BR 87.4 puts before an override.
func assertDiscountable(
	ctx corectx.Context, order dmodel.DynamicFields, params GrantManualDiscountParams,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	// The reason is a business invariant, not an access decision, so it is enforced here rather
	// than in app/. An override with no stated cause is indistinguishable from a mistake, and it is
	// the field an auditor asking why this customer paid less actually reads.
	if params.Reason == "" {
		vErrs.Append(*ft.NewBusinessViolation("reason", ReasonDiscountReasonRequired,
			"a manual discount must say why the price was changed"))
	}

	if !params.Amount.IsPositive() {
		vErrs.Append(*ft.NewBusinessViolation("discount_amount", ReasonDiscountNotPositive,
			"a manual discount must be a positive amount; a negative one would be a surcharge"))
	}

	// A CONFIRMED order is frozen (BR 11). Discounting one would change what a customer already
	// agreed to pay, after any bill was raised against it — so the correction after confirmation is
	// a return or a refund, each with its own money movement, not a retrospective price edit.
	status := stringOf(order, models.SalesOrderFieldStatus)
	if status != string(models.SalesOrderStatusDraft) {
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonOrderNotDiscountable,
			"an order in status '"+status+"' is frozen; correct it with a return or a refund"))
	}

	// A line-level override must name a line of THIS order. Naming another order's line would move
	// an approval onto a sale nobody authorised it for, and the engine — which is pure and takes its
	// input on trust — would silently ignore it, leaving an override that appears granted and does
	// nothing.
	if params.SalesOrderLineId != "" {
		line, err := loadRecord(ctx, models.SalesOrderLineSchemaName,
			models.SalesOrderLineFieldId, params.SalesOrderLineId)
		if err != nil {
			return nil, err
		}
		if line == nil ||
			stringOf(line, models.SalesOrderLineFieldSalesOrderId) != params.SalesOrderId {
			vErrs.Append(*ft.NewBusinessViolation("sales_order_line_id",
				ReasonDiscountLineNotFound,
				"line '"+params.SalesOrderLineId+"' does not belong to this order"))
		}
	}

	if vErrs.Count() > 0 {
		return vErrs, nil
	}
	return nil, nil
}

// writeManualDiscount stores the override and audits it.
//
// Both writes in ONE transaction. BR 87.4 requires the old and the new price to be audited, and an
// override stored without its audit entry is exactly the untraceable price change the requirement
// exists to prevent.
func writeManualDiscount(
	ctx corectx.Context,
	order dmodel.DynamicFields,
	params GrantManualDiscountParams,
	totalBefore decimal.Decimal,
) (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", err
	}
	discountId := string(*id)
	orgId := stringOf(order, basemodel.FieldOrgId)

	err = withTransaction(ctx, models.SalesManualDiscountSchemaName,
		func(tranxCtx corectx.Context) error {
			engine, err := engineFor(models.SalesManualDiscountSchemaName)
			if err != nil {
				return err
			}

			record := dmodel.DynamicFields{
				models.SalesManualDiscountFieldId:           discountId,
				models.SalesManualDiscountFieldSalesOrderId: params.SalesOrderId,
				models.SalesManualDiscountFieldAmount:       params.Amount,
				models.SalesManualDiscountFieldReason:       params.Reason,

				// The price as it stood when the operator decided. Stored rather than derived,
				// because repricing moves the surrounding numbers afterwards and this is the only
				// record of what they were actually looking at.
				models.SalesManualDiscountFieldOriginalAmount: totalBefore,

				basemodel.FieldOrgId: orgId,
			}
			if params.SalesOrderLineId != "" {
				record[models.SalesManualDiscountFieldOrderLineId] = params.SalesOrderLineId
			}
			if actorId := salesActorOf(ctx); actorId != "" {
				record[models.SalesManualDiscountFieldGrantedBy] = actorId
			}

			if _, err := engine.ResourceRepository().Insert(tranxCtx, record); err != nil {
				return err
			}

			// BR 87.4's audit requirement: both prices, and the reason, on the order's own trail.
			// The ACTOR is not passed: WriteSalesAuditEvent takes it from the context, so that an
			// operation cannot record somebody else as having performed it. granted_by on the row
			// above is the caller's own record of who authorised it, which is a different claim.
			return WriteSalesAuditEvent(tranxCtx, SalesAuditEntry{
				SalesOrderId: params.SalesOrderId,
				EntityType:   models.SalesManualDiscountSchemaName,
				EntityId:     discountId,
				Action:       models.SalesOrderActionManualDiscount,
				Reason:       params.Reason,
				OrgId:        orgId,
				Metadata: map[string]any{
					"original_amount":     totalBefore,
					"discount_amount":     params.Amount,
					"sales_order_line_id": params.SalesOrderLineId,
				},
			})
		})
	if err != nil {
		return "", err
	}
	return discountId, nil
}

// RevokeManualDiscount removes an override and reprices.
//
// A hard delete rather than a status, and deliberately: the audit trail already records that the
// discount was granted and by whom, so the history survives the row — while a revoked-but-present
// row would have to be filtered out of the engine input by every future reader, and the one that
// forgot would reapply a discount somebody had withdrawn.
func RevokeManualDiscount(
	ctx corectx.Context,
	orderId, discountId string,
	taxSvc itExt.TaxCalculationExtService,
	policy SalesPolicy,
) (*ft.ClientErrors, error) {
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return OrderNotFoundErrors(orderId), nil
	}

	status := stringOf(order, models.SalesOrderFieldStatus)
	if status != string(models.SalesOrderStatusDraft) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("status", ReasonOrderNotDiscountable,
			"an order in status '"+status+"' is frozen"))
		return vErrs, nil
	}

	err = withTransaction(ctx, models.SalesManualDiscountSchemaName,
		func(tranxCtx corectx.Context) error {
			engine, err := engineFor(models.SalesManualDiscountSchemaName)
			if err != nil {
				return err
			}
			if _, err := engine.ResourceRepository().DeleteOne(tranxCtx, dmodel.DynamicFields{
				models.SalesManualDiscountFieldId: discountId,
			}); err != nil {
				return err
			}

			return WriteSalesAuditEvent(tranxCtx, SalesAuditEntry{
				SalesOrderId: orderId,
				EntityType:   models.SalesManualDiscountSchemaName,
				EntityId:     discountId,
				Action:       models.SalesOrderActionManualDiscount,
				Reason:       "manual discount revoked",
				OrgId:        stringOf(order, basemodel.FieldOrgId),
			})
		})
	if err != nil {
		return nil, err
	}

	_, vErrs, err := RepriceOrder(ctx, orderId, taxSvc, policy)
	return vErrs, err
}
