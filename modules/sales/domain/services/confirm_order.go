package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	lock "github.com/sky-as-code/nikki-erp/modules/core/infra/distributedlock"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// Confirming a sales order: the moment the business commits. Before it a draft may be repriced
// freely, after it the numbers are what the customer was promised.
//
// The ORDER of the steps is the substance:
//
// 	1. validate the order is confirmable
// 	2. reprice - catalogue prices may have moved since the draft was last touched
// 	3. redeem the voucher reservations, failing the confirm if one was exhausted meanwhile
// 	4. freeze the snapshot and stamp confirmed_at
// 	5. raise a fulfilment request for stock-managed lines
// 	6. create the initial bill                              [not built]
//
// Step 6 is absent rather than stubbed: a placeholder that silently did nothing would make a
// confirm look complete when it is not. See ConfirmOrderResult.Pending.
//
// It all runs under a distributed lock: confirm is not a single-row update, so the etag that
// guards ordinary edits cannot protect it, and a second confirm interleaved anywhere could redeem
// the same voucher twice. KNOWN LIMITATION of that lock (see
// vending_machine_new/domain/services/order_lock.go): Release deletes the key unconditionally with
// no fencing token, so an operation overrunning the TTL deletes a second caller's lock. Nothing
// under the lock may therefore block unboundedly - hence the bounded reprice and no gateway call.

const (
	// confirmLockTtl must outlast the slowest thing done under it, the reprice: one tax call plus a
	// handful of row writes.
	confirmLockTtl = 20 * time.Second

	// A contended order is worth waiting for: the competing holder is normally another confirm that
	// finishes in well under a second.
	confirmLockRetryCount = 5
	confirmLockRetryDelay = 500 * time.Millisecond
)

type ConfirmOrderResult struct {
	SalesOrderId string
	Status       string
	ConfirmedAt  string

	Pricing *RepriceResult

	// RedeemedVoucherIds are the reservations turned into real uses by this confirm.
	RedeemedVoucherIds []string

	// Fulfillment is what Inventory answered. Nil when the order needed no fulfilment at all - every
	// line a service or a digital item - which is legitimate, so a caller that must know whether goods
	// are coming reads this rather than assuming a reservation exists.
	Fulfillment *RaiseFulfillmentResult

	// Pending names the steps this confirm did not complete.
	Pending []string
}

// The refusal reasons confirm can produce.
const (
	ReasonNotConfirmable   = "sales_order.not_confirmable"
	ReasonEmptyOrder       = "sales_order.no_lines"
	ReasonChannelMissing   = "sales_order.sales_channel_missing"
	ReasonVoucherExhausted = "sales_order.voucher_exhausted"
	ReasonLockUnavailable  = "sales_order.locked"
)

// ConfirmOrder commits a draft order. Returns ClientErrors for every refusal a caller could fix; a
// contended lock is one of them, since the caller may simply try again.
func ConfirmOrder(
	ctx corectx.Context,
	orderId string,
	dLock lock.DistributedLock,
	taxSvc itExt.TaxCalculationExtService,
	fulfillment itExt.FulfillmentExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	if dLock == nil {
		// Confirming without the lock would lose the only protection against a double redemption.
		return nil, nil, errors.New(
			"the distributed lock is not available; a sales order cannot be confirmed without it")
	}

	key := confirmLockKeyOf(orderId)
	acquired, err := dLock.AcquireWithRetry(
		ctx, key, confirmLockTtl, confirmLockRetryCount, confirmLockRetryDelay)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "acquiring the lock of order '%s'", orderId)
	}
	if !acquired {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("id", ReasonLockUnavailable,
			"this order is being changed by another request; try again"))
		return nil, vErrs, nil
	}
	defer func() {
		// Best-effort: a failure leaves a key that expires on its own TTL, a delay not a corruption.
		_ = dLock.Release(ctx, key)
	}()

	// Read AFTER acquiring the lock: a record read while queuing describes the world as it was
	// before the other holder finished.
	return confirmUnderLock(ctx, orderId, taxSvc, fulfillment, basisSvc, policy)
}

// confirmLockKeyOf builds the key from the id rather than the order number, which is a
// human-facing label a correction could reuse.
func confirmLockKeyOf(orderId string) string {
	return "lock:sales_order:" + orderId
}

func confirmUnderLock(
	ctx corectx.Context,
	orderId string,
	taxSvc itExt.TaxCalculationExtService,
	fulfillment itExt.FulfillmentExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	if vErrs, err := assertConfirmable(ctx, record); err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Step 2: reprice, because a catalogue price may have moved since the draft was last touched.
	priced, vErrs, err := RepriceOrder(ctx, orderId, taxSvc, policy, basisSvc)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Step 3: redeem the vouchers before the status moves, so a voucher exhausted meanwhile fails
	// the confirm with the order still a draft - editable, and confirmable again with another code.
	redeemed, vErrs, err := redeemOrderVouchers(ctx, orderId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Step 4: freeze. The snapshot columns become immutable the moment the status is `confirmed`,
	// enforced by assertSnapshotsUnchanged on every later write.
	confirmedAt := time.Now().UTC()
	if err := stampConfirmed(ctx, orderId, record, confirmedAt); err != nil {
		return nil, nil, err
	}

	// Step 5: ask Inventory to hold the goods, AFTER the order is committed and outside its
	// transaction, since the two cannot be atomic. A confirmed order with no reservation is
	// recoverable, whereas a reservation against an order that failed to confirm holds stock nothing
	// will claim. A refusal does NOT fail the confirm.
	fulfilment, vErrs, err := RaiseFulfillmentRequest(
		ctx, orderId, string(models.SalesFulfillmentTypeReservation), fulfillment)
	if err != nil {
		return nil, nil, err
	}

	return &ConfirmOrderResult{
		SalesOrderId:       orderId,
		Status:             string(models.SalesOrderStatusConfirmed),
		ConfirmedAt:        confirmedAt.Format(time.RFC3339),
		Pricing:            priced,
		RedeemedVoucherIds: redeemed,
		Fulfillment:        fulfilment,
		Pending:            pendingConfirmSteps(fulfilment, vErrs),
	}, nil, nil
}

// pendingConfirmSteps is returned to the caller rather than logged, because a kiosk that believes
// a confirm was complete will happily dispense goods against an order with no bill.
func pendingConfirmSteps(result *RaiseFulfillmentResult, vErrs *ft.ClientErrors) []string {
	pending := make([]string, 0, 2)

	switch {
	case vErrs != nil:
		// Nothing to fulfil is the ordinary case and is reported rather than hidden, so a kiosk does
		// not wait for a delivery that is not coming.
		pending = append(pending, "fulfilment_request ("+describeFirstViolation(vErrs)+")")
	case result == nil:
		pending = append(pending, "fulfilment_request (no request was raised)")
	case !result.Dispatched:
		pending = append(pending, "fulfilment_request (no inventory port bound)")
	case result.Status == string(models.SalesFulfillmentStatusRejected):
		pending = append(pending, "fulfilment_request (inventory rejected the request)")
	}

	// The bill itself is raised later, not by confirm. Reported rather than omitted for the same
	// reason as the fulfilment step: a caller told the confirm succeeded would otherwise assume
	// there is something to take payment against.
	pending = append(pending, "initial_bill (confirm does not raise a bill)")
	return pending
}

func describeFirstViolation(vErrs *ft.ClientErrors) string {
	if vErrs == nil || len(*vErrs) == 0 {
		return "no reason given"
	}
	first := (*vErrs)[0]
	if first.Message != "" {
		return first.Message
	}
	return first.Key
}

func assertConfirmable(
	ctx corectx.Context, record dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	refuse := func(field, reason, message string) *ft.ClientErrors {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(field, reason, message))
		return vErrs
	}

	status := stringOf(record, models.SalesOrderFieldStatus)
	if status != string(models.SalesOrderStatusDraft) {
		// Re-confirming is refused rather than a no-op: a confirm has side effects - it redeems
		// vouchers and raises a fulfilment request - so a silent second success would redeem twice.
		return refuse("status", ReasonNotConfirmable,
			"only a draft sales order may be confirmed; this one is '"+status+"'"), nil
	}

	// The check stays even though the column is NOT NULL: a schema that changes underneath this
	// code should fail loudly here rather than silently later.
	if stringOf(record, models.SalesOrderFieldSalesChannelId) == "" {
		return refuse("sales_channel_id", ReasonChannelMissing,
			"this order has no sales channel and cannot be confirmed"), nil
	}

	orderId := stringOf(record, models.SalesOrderFieldId)
	lines, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		// An empty draft is a valid draft but not a valid sale.
		return refuse("lines", ReasonEmptyOrder,
			"a sales order with no lines cannot be confirmed"), nil
	}
	return nil, nil
}

// redeemOrderVouchers turns this order's reservations into real uses. One that can no longer be
// redeemed fails the WHOLE confirm: the customer was quoted a price including the discount, so
// confirming without it would overcharge them. Failing leaves the order a draft.
func redeemOrderVouchers(
	ctx corectx.Context, orderId string,
) ([]string, *ft.ClientErrors, error) {
	redemptions, err := searchBy(ctx,
		models.SalesVoucherRedemptionSchemaName,
		models.SalesVoucherRedemptionFieldSalesOrderId, orderId)
	if err != nil {
		return nil, nil, err
	}

	redeemed := make([]string, 0, len(redemptions))
	for _, record := range redemptions {
		status := stringOf(record, models.SalesVoucherRedemptionFieldStatus)
		if status != string(models.VoucherRedemptionStatusReserved) {
			// Already redeemed, released or reversed. Not this confirm's business: a released
			// reservation was deliberately given back, and a redeemed one means a retry.
			continue
		}

		redemptionId := stringOf(record, models.SalesVoucherRedemptionFieldId)
		vErrs, err := SettleRedemption(ctx, redemptionId,
			string(models.VoucherRedemptionStatusRedeemed))
		if err != nil {
			return nil, nil, err
		}
		if vErrs != nil {
			// Reported as an exhausted voucher, and the confirm is abandoned with the order still a
			// draft.
			out := ft.NewClientErrors()
			out.Append(*ft.NewBusinessViolation("voucher", ReasonVoucherExhausted,
				"a voucher applied to this order can no longer be redeemed; "+
					"remove it and try again"))
			return nil, out, nil
		}
		redeemed = append(redeemed,
			stringOf(record, models.SalesVoucherRedemptionFieldVoucherCodeId))
	}
	return redeemed, nil, nil
}

// stampConfirmed moves the status and records when. Fulfilment status is derived here rather than
// left at its default, which stops a fully-digital sale waiting forever on a warehouse.
//
// The three writes are ONE transaction: an integration event announcing a confirmation that then
// rolled back is exactly the failure the outbox exists to prevent.
func stampConfirmed(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields, at time.Time,
) error {
	return withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		return stampConfirmedInTranx(tranxCtx, orderId, record, at)
	})
}

func stampConfirmedInTranx(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields, at time.Time,
) error {
	engine, err := engineFor(models.SalesOrderSchemaName)
	if err != nil {
		return err
	}

	update := dmodel.DynamicFields{
		models.SalesOrderFieldId:          orderId,
		models.SalesOrderFieldStatus:      string(models.SalesOrderStatusConfirmed),
		models.SalesOrderFieldConfirmedAt: model.ModelDateTime(at),
	}

	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return err
	}

	orgId := stringOf(record, basemodel.FieldOrgId)
	if err := WriteOrderStatusEvent(ctx, orderId, models.SalesOrderActionConfirm,
		string(models.SalesOrderStatusDraft), string(models.SalesOrderStatusConfirmed),
		orgId); err != nil {
		return err
	}

	// The INTEGRATION event, distinct from the audit event just written: that one is Sales' own
	// history, this one a public announcement. Both are written here so neither can exist without
	// the status change that justifies it.
	_, err = RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesOrderConfirmed,
		AggregateId: orderId,
		OrgId:       orgId,
		OccurredAt:  at.Unix(),
		Payload: map[string]any{
			"sales_order_id": orderId,
			"order_number":   stringOf(record, models.SalesOrderFieldOrderNumber),
			"currency_code":  stringOf(record, models.SalesOrderFieldCurrencyCode),

			// The totals travel WITH the event so a consumer never reads back into Sales, and acts on
			// what was true at confirmation.
			"grand_total": decimalOf(record, models.SalesOrderFieldGrandTotal),
			"tax_total":   decimalOf(record, models.SalesOrderFieldTaxTotal),

			"confirmed_at": at.Unix(),
		},
	})
	return err
}
