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
// 	4. freeze the snapshot, create the initial bill, stamp confirmed_at and announce the stage
// 	5. raise a fulfilment request for stock-managed lines
//
// Step 4 is ONE transaction, and the bill is part of it deliberately: an order that says confirmed
// with no bill is a sale nobody can take money against, and it is the state the whole operation is
// arranged to make impossible. The stage event goes into the outbox in that same transaction, so a
// consumer told the order is confirmed can rely on the bill already being there.
//
// A second confirm of an already-confirmed order is NOT an error: it returns the original bill
// without creating another. See assertConfirmable.
//
// It all runs under a distributed lock: confirm is not a single-row update, so the etag that
// guards ordinary edits cannot protect it, and a second confirm interleaved anywhere could redeem
// the same voucher twice. KNOWN LIMITATION of that lock (see
// vendingmachine/domain/services/order_lock.go): Release deletes the key unconditionally with
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

	// InitialBillId is the bill this confirmation raised, or the one the FIRST confirmation raised
	// when this call was a retry. Always set on success: a confirmed order without a bill is a
	// data-integrity fault, not a state a caller has to handle.
	InitialBillId string

	// AlreadyConfirmed says this call found the work already done and changed nothing. The caller
	// still gets the order and its initial bill, which is the point: a client that lost the first
	// response can ask again without risking a second bill.
	AlreadyConfirmed bool

	Pricing *RepriceResult

	// RedeemedVoucherIds are the reservations turned into real uses by this confirm.
	RedeemedVoucherIds []string

	// Fulfillment is what Inventory answered. Nil when the order needed no fulfilment at all - every
	// line a service or a digital item - which is legitimate, so a caller that must know whether goods
	// are coming reads this rather than assuming a reservation exists.
	Fulfillment *RaiseFulfillmentResult

	// KioskFulfillment is the delivery half of a kiosk sale: the fulfillment created and the stock
	// held for it, before any money moved. Nil for every other kind of order, which is how a caller
	// tells "this sale dispenses from a machine" from "this sale ships".
	KioskFulfillment *CreateFulfillmentResult

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

// ConfirmOrderOptions carries what a confirm records beyond the order id.
type ConfirmOrderOptions struct {
	// ConfirmationNote is the confirmer's note, optional. Automatic overrides it with exactly
	// AutoConfirmedNote, so the audit trail reads the same for every system confirm.
	ConfirmationNote string
	Automatic        bool
}

// ConfirmOrder commits a draft order with no note. See ConfirmOrderWith.
func ConfirmOrder(
	ctx corectx.Context,
	orderId string,
	dLock lock.DistributedLock,
	taxSvc itExt.TaxCalculationExtService,
	fulfillment itExt.FulfillmentExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
	methods *SalesFulfillmentMethodDomainServiceImpl,
	reservations itExt.FulfillmentReservationExtService,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	return ConfirmOrderWith(ctx, orderId, ConfirmOrderOptions{}, dLock, taxSvc, fulfillment, basisSvc, policy, methods, reservations)
}

// ConfirmOrderWith commits a draft order. Returns ClientErrors for every refusal a caller could
// fix; a contended lock is one of them, since the caller may simply try again. Manual and
// automatic confirms run exactly the same validation and steps; only the note differs.
func ConfirmOrderWith(
	ctx corectx.Context,
	orderId string,
	opts ConfirmOrderOptions,
	dLock lock.DistributedLock,
	taxSvc itExt.TaxCalculationExtService,
	fulfillment itExt.FulfillmentExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
	methods *SalesFulfillmentMethodDomainServiceImpl,
	reservations itExt.FulfillmentReservationExtService,
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
	return confirmUnderLock(ctx, orderId, opts, taxSvc, fulfillment, basisSvc, policy, methods, reservations)
}

// confirmLockKeyOf builds the key from the id rather than the order number, which is a
// human-facing label a correction could reuse.
func confirmLockKeyOf(orderId string) string {
	return "lock:sales_order:" + orderId
}

func confirmUnderLock(
	ctx corectx.Context,
	orderId string,
	opts ConfirmOrderOptions,
	taxSvc itExt.TaxCalculationExtService,
	fulfillment itExt.FulfillmentExtService,
	basisSvc itExt.ProductPricingBasisExtService,
	policy SalesPolicy,
	methods *SalesFulfillmentMethodDomainServiceImpl,
	reservations itExt.FulfillmentReservationExtService,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, nil, err
	}
	if record == nil {
		return nil, OrderNotFoundErrors(orderId), nil
	}

	// A confirm that already happened answers with what it produced rather than refusing. The
	// client that lost the response is the ordinary case, and the alternative - an error - pushes it
	// towards creating a second order for a sale that is already on the books.
	replay, vErrs, err := replayConfirm(record)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	if replay != nil {
		return replay, nil, nil
	}

	if vErrs, err := assertConfirmable(ctx, record); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	// An order past its payment deadline may only be cancelled: confirming it would hold stock
	// for a customer who can no longer pay.
	if vErrs, err := assertOrderNotExpired(ctx, record); err != nil || vErrs != nil {
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

	// Step 3b: for a kiosk sale ONLY, create the fulfillment and hold its stock BEFORE the order is
	// frozen. This inverts the ordering of step 5 below, and does so deliberately: a customer
	// standing at a machine has no way to work through an unfulfillable order afterwards, so a
	// reservation that cannot be met must refuse here, with the order still a draft and nothing
	// charged. Every other kind of order skips this entirely and keeps the original sequence.
	kioskFulfillment, vErrs, err := CreateKioskFulfillment(ctx, record, methods, reservations)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Steps 4 and 6: freeze, and raise the initial bill. The snapshot columns become immutable the
	// moment the status is `confirmed`, enforced by assertSnapshotsUnchanged on every later write.
	// The bill, the link to it, the status and the stage event are ONE transaction: a confirmed
	// order with no bill is a sale nobody can pay, and an announcement of a confirmation that then
	// rolled back is what the outbox exists to prevent.
	confirmedAt := time.Now().UTC()
	note := normaliseNote(opts.ConfirmationNote)
	if opts.Automatic {
		note = models.AutoConfirmedNote
	}
	initialBillId, err := stampConfirmed(ctx, orderId, record, confirmedAt, policy, note)
	if err != nil {
		return nil, nil, err
	}

	// Step 5: ask Inventory to hold the goods, AFTER the order is committed and outside its
	// transaction, since the two cannot be atomic. A confirmed order with no reservation is
	// recoverable, whereas a reservation against an order that failed to confirm holds stock nothing
	// will claim. A refusal does NOT fail the confirm.
	// Skipped for a kiosk sale: its goods were already claimed at step 3b, and raising the intent
	// request as well would hold the same stock twice — once against the fulfillment and once
	// against a second transfer nothing will ever release.
	var fulfilment *RaiseFulfillmentResult
	if kioskFulfillment == nil {
		fulfilment, vErrs, err = RaiseFulfillmentRequest(
			ctx, orderId, string(models.SalesFulfillmentTypeReservation), fulfillment)
		if err != nil {
			return nil, nil, err
		}
	}

	return &ConfirmOrderResult{
		SalesOrderId:       orderId,
		Status:             string(models.SalesOrderStatusConfirmed),
		ConfirmedAt:        confirmedAt.Format(time.RFC3339),
		InitialBillId:      initialBillId,
		Pricing:            priced,
		RedeemedVoucherIds: redeemed,
		Fulfillment:        fulfilment,
		KioskFulfillment:   kioskFulfillment,
		Pending:            pendingConfirmSteps(fulfilment, vErrs, kioskFulfillment),
	}, nil, nil
}

// pendingConfirmSteps is returned to the caller rather than logged, because a kiosk that believes
// a confirm was complete will happily dispense goods against an order with no bill.
func pendingConfirmSteps(
	result *RaiseFulfillmentResult, vErrs *ft.ClientErrors, kiosk *CreateFulfillmentResult,
) []string {
	pending := make([]string, 0, 2)

	// A kiosk sale reports on its own reservation instead: it took one at step 3b and raised no
	// intent request, so describing the absent request as pending would name a step that was never
	// going to run.
	if kiosk != nil {
		if kiosk.Status != string(models.FulfillmentStatusReserved) {
			pending = append(pending, "fulfillment_reservation (the stock is not held)")
		}
		return pending
	}

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

// replayConfirm answers a repeated confirm with what the first one produced, or nil when this is a
// genuine first attempt.
//
// Only `confirmed` replays. A `processing`, `completed` or `cancelled` order has moved on since its
// confirmation, so a caller asking to confirm it is not retrying a lost response but acting on a
// stale view - assertConfirmable refuses those, as before.
//
// The bill is deliberately NOT re-derived here: initial_bill_id names the bill the FIRST confirm
// raised, and searching for a current one would answer with whatever a later split or merge left
// open, which is a different question.
func replayConfirm(record dmodel.DynamicFields) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	if stringOf(record, models.SalesOrderFieldStatus) != string(models.SalesOrderStatusConfirmed) {
		return nil, nil, nil
	}

	orderId := stringOf(record, models.SalesOrderFieldId)
	billId := stringOf(record, models.SalesOrderFieldInitialBillId)
	if billId == "" {
		// Confirmed with no bill: an invariant violation, and NOT something to repair by raising one
		// now. The order may have been billed and paid through a path that predates this column, and
		// a fresh bill would ask the customer for money they have already handed over. It is a
		// migration's job to work out which (CR §25), with the evidence this code does not have.
		return nil, nil, errors.New("sales order '" + orderId +
			"' is confirmed but names no initial bill; this is a data-integrity fault and must be " +
			"reconciled rather than re-confirmed")
	}

	confirmedAt := ""
	if at := dateTimeOf(record, models.SalesOrderFieldConfirmedAt); at != nil {
		confirmedAt = at.GoTime().Format(time.RFC3339)
	}

	return &ConfirmOrderResult{
		SalesOrderId:     orderId,
		Status:           string(models.SalesOrderStatusConfirmed),
		ConfirmedAt:      confirmedAt,
		InitialBillId:    billId,
		AlreadyConfirmed: true,

		// Nothing is repriced, redeemed or reserved on a replay, so those stay empty rather than
		// restating work this call did not do.
		Pending: []string{},
	}, nil, nil
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
	policy SalesPolicy, note string,
) (billId string, err error) {
	err = withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		billId, err = stampConfirmedInTranx(tranxCtx, orderId, record, at, policy, note)
		return err
	})
	if err != nil {
		return "", err
	}
	return billId, nil
}

func stampConfirmedInTranx(
	ctx corectx.Context, orderId string, record dmodel.DynamicFields, at time.Time,
	policy SalesPolicy, note string,
) (string, error) {
	engineRepo, err := repoFor(models.SalesOrderSchemaName)
	if err != nil {
		return "", err
	}

	// The bill FIRST, so the order can be written already pointing at it: the link and the status
	// are one write, and no committed state ever has a confirmed order whose initial_bill_id is
	// still being filled in.
	billId, err := CreateInitialBill(ctx, record, policy)
	if err != nil {
		return "", err
	}

	update := dmodel.DynamicFields{
		models.SalesOrderFieldId:            orderId,
		models.SalesOrderFieldStatus:        string(models.SalesOrderStatusConfirmed),
		models.SalesOrderFieldConfirmedAt:   model.ModelDateTime(at),
		models.SalesOrderFieldInitialBillId: billId,
	}
	if note != "" {
		update[models.SalesOrderFieldConfirmationNote] = note
	}

	if _, err := engineRepo.Update(ctx, update); err != nil {
		return "", err
	}

	// Checked after the write and inside the transaction, the way a split checks itself: the
	// allocations of a bill mean nothing until they are all written, and an initial bill that does
	// not add up to its order must take the whole confirmation down with it.
	vErrs, err := AssertOrderAllocationBalances(ctx, orderId)
	if err != nil {
		return "", err
	}
	if vErrs != nil {
		return "", errors.New("the initial bill did not balance against its order: " +
			vErrs.ToError().Error())
	}

	orgId := stringOf(record, basemodel.FieldOrgId)
	if err := WriteOrderStatusEvent(ctx, orderId, models.SalesOrderActionConfirm,
		string(models.SalesOrderStatusDraft), string(models.SalesOrderStatusConfirmed),
		orgId); err != nil {
		return "", err
	}

	// The INTEGRATION event, distinct from the audit event just written: that one is Sales' own
	// history, this one a public announcement. Both are written here so neither can exist without
	// the status change that justifies it.
	//
	// initial_bill_id travels with it: a consumer told an order is confirmed may rely on the bill
	// already existing in committed state, which is only true because both were written here.
	err = RecordOrderStageChanged(ctx, OrderStageChangedParams{
		Order:         record,
		PreviousStage: string(models.SalesOrderStatusDraft),
		CurrentStage:  string(models.SalesOrderStatusConfirmed),
		StageVersion:  nextStageVersion(record),
		InitialBillId: billId,
		OccurredAt:    at,
	})
	if err != nil {
		return "", err
	}
	return billId, nil
}
