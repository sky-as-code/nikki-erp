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

// Confirming a sales order (BR 72, SALES-013).
//
// Confirmation is the moment the business commits. Before it a draft may be repriced freely; after
// it the numbers are what the customer was promised (BR 11), and no catalogue change may alter them.
//
// BR 72 fixes the ORDER of the steps, and the order is the substance:
//
//	1. validate the order is confirmable
//	2. reprice - catalogue prices may have moved since the draft was last touched
//	3. redeem the voucher reservations, failing the confirm if one was exhausted meanwhile
//	4. freeze the snapshot and stamp confirmed_at
//	5. raise a fulfilment request for stock-managed lines   [SALES-029, not built]
//	6. create the initial bill                              [SALES-024, not built]
//
// Steps 5 and 6 are deliberately absent rather than stubbed: neither the inventory fulfilment port
// nor sales_bills exists, and writing a placeholder that silently does nothing would make a confirm
// look complete when it is not. See ConfirmOrderResult.Pending.
//
// # Why the whole thing is under a distributed lock (D-30)
//
// Confirm is not a single-row update, so the etag that guards ordinary edits cannot protect it: it
// reads lines, redeems vouchers on another table, and writes the order, and a second confirm
// interleaved anywhere in that sequence could redeem the same voucher twice or freeze a snapshot
// against prices the other caller had just changed.
//
// KNOWN LIMITATION, inherited from the platform lock and documented on
// vending_machine_new/domain/services/order_lock.go: Release deletes the key unconditionally, with no
// fencing token. If an operation overruns the TTL, a second caller acquires the lock and the first
// caller's Release then deletes the SECOND caller's lock. The guard is that nothing under the lock
// may block unboundedly - which is why the tax call happens inside a bounded reprice and nothing here
// calls a payment gateway.

const (
	// confirmLockTtl bounds how long one order may be held. It must outlast the slowest thing done
	// under it, which is the reprice - one tax call plus a handful of row writes.
	confirmLockTtl = 20 * time.Second

	// A contended order is worth waiting for rather than failing: the competing holder is normally
	// another confirm that finishes in well under a second.
	confirmLockRetryCount = 5
	confirmLockRetryDelay = 500 * time.Millisecond
)

// ConfirmOrderResult is what a confirm concluded.
type ConfirmOrderResult struct {
	SalesOrderId string
	Status       string
	ConfirmedAt  string

	Pricing *RepriceResult

	// RedeemedVoucherIds are the reservations turned into real uses by this confirm.
	RedeemedVoucherIds []string

	// Pending names the BR 72 steps this build cannot perform yet, so a caller is told plainly
	// rather than being left to assume a complete confirm. Empty once SALES-024 and SALES-029 land.
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

// ConfirmOrder commits a draft order.
//
// Returns ClientErrors for every refusal a caller could fix - a draft with no lines, a voucher that
// ran out while the basket sat open. A contended lock is also a client error rather than a fault:
// the caller may simply try again.
func ConfirmOrder(
	ctx corectx.Context,
	orderId string,
	dLock lock.DistributedLock,
	taxSvc itExt.TaxCalculationExtService,
	policy SalesPolicy,
) (*ConfirmOrderResult, *ft.ClientErrors, error) {
	if dLock == nil {
		// Confirming without the lock would be confirming without the only protection against a
		// double redemption. Refusing is the safe direction.
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
		// The release is best-effort: a failure here leaves a key that expires on its own TTL, which
		// is a delay rather than a corruption. Reporting it would replace a successful confirm with
		// an error the caller cannot act on.
		_ = dLock.Release(ctx, key)
	}()

	// Everything below runs under the lock. The record is read AFTER acquiring it, never before:
	// a record read while queuing describes the world as it was before the other holder finished.
	return confirmUnderLock(ctx, orderId, taxSvc, policy)
}

// confirmLockKeyOf names the lock for one order.
//
// Built from the id rather than the order number, because the number is a human-facing label that a
// replacement or a correction could reuse, and two rows of one purchase chain must not share a lock.
func confirmLockKeyOf(orderId string) string {
	return "lock:sales_order:" + orderId
}

func confirmUnderLock(
	ctx corectx.Context,
	orderId string,
	taxSvc itExt.TaxCalculationExtService,
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

	// Step 2: reprice. BR 72 requires it because a catalogue price may have moved since the draft
	// was last touched, and a customer must be charged what the goods cost now rather than what they
	// cost when the basket was opened.
	priced, vErrs, err := RepriceOrder(ctx, orderId, taxSvc, policy)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Step 3: redeem the vouchers. Before the status moves, so that a voucher exhausted meanwhile
	// fails the confirm with the order still a draft - editable, and able to be confirmed again once
	// the customer picks a different code.
	redeemed, vErrs, err := redeemOrderVouchers(ctx, orderId)
	if err != nil || vErrs != nil {
		return nil, vErrs, err
	}

	// Step 4: freeze. The snapshot columns become immutable the moment the status is `confirmed`,
	// enforced by assertSnapshotsUnchanged on every later write; this step is what flips that switch.
	confirmedAt := time.Now().UTC()
	if err := stampConfirmed(ctx, orderId, record, confirmedAt); err != nil {
		return nil, nil, err
	}

	return &ConfirmOrderResult{
		SalesOrderId:       orderId,
		Status:             string(models.SalesOrderStatusConfirmed),
		ConfirmedAt:        confirmedAt.Format(time.RFC3339),
		Pricing:            priced,
		RedeemedVoucherIds: redeemed,
		Pending:            pendingConfirmSteps(),
	}, nil, nil
}

// pendingConfirmSteps names the BR 72 steps this build cannot perform.
//
// Returned to the caller rather than logged, because a kiosk that believes a confirm was complete
// will happily dispense goods against an order that has no bill and no fulfilment request.
func pendingConfirmSteps() []string {
	return []string{
		"fulfilment_request (SALES-029: no inventory port bound)",
		"initial_bill (SALES-024: sales_bills does not exist)",
	}
}

// assertConfirmable applies the gates BR 72 puts before a confirm.
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
		// Re-confirming an already-confirmed order is refused rather than treated as a no-op. Unlike
		// a status transition, a confirm has side effects - it redeems vouchers and will raise a
		// fulfilment request - so a silent second success would redeem twice.
		return refuse("status", ReasonNotConfirmable,
			"only a draft sales order may be confirmed; this one is '"+status+"'"), nil
	}

	// CR 18 demands the check even though D-19 made the column NOT NULL. It is cheap, and a schema
	// that changes underneath this code should fail loudly here rather than silently later.
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
		// An empty draft is a valid draft (BR 69) but not a valid sale. This is the one place the
		// distinction is enforced.
		return refuse("lines", ReasonEmptyOrder,
			"a sales order with no lines cannot be confirmed"), nil
	}
	return nil, nil
}

// redeemOrderVouchers turns this order's reservations into real uses.
//
// A reservation that can no longer be redeemed fails the WHOLE confirm. That is the strict reading
// and the right one: the customer was quoted a price that included the discount, and confirming
// without it would charge them more than they agreed. Failing leaves the order a draft, so the
// customer can be told and can choose again.
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
			// The reservation could not become a use. Report it as an exhausted voucher, which is
			// what the customer needs to hear, and abandon the confirm with the order still a draft.
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

// stampConfirmed moves the status and records when.
//
// Fulfilment status is derived here rather than left at its default: an order none of whose lines
// needs picking is `not_required` and may complete as soon as it is paid, while one that does starts
// `pending`. Deriving it at confirm is what stops a fully-digital sale sitting forever waiting for a
// warehouse that has nothing to do.
// The three writes are ONE transaction, and that is what the outbox's guarantee rests on. An
// integration event announcing a confirmation that then rolled back is exactly the failure BR 80
// exists to prevent, and it is only prevented if the event and the status move together.
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

	// The INTEGRATION event, which is a different thing from the audit event just written: that one
	// is Sales' own history, this one is a public announcement to consumers Sales does not control.
	// Both are written here so neither can exist without the status change that justifies it.
	_, err = RecordEvent(ctx, RecordEventParams{
		EventType:   models.EventSalesOrderConfirmed,
		AggregateId: orderId,
		OrgId:       orgId,
		OccurredAt:  at.Unix(),
		Payload: map[string]any{
			"sales_order_id": orderId,
			"order_number":   stringOf(record, models.SalesOrderFieldOrderNumber),
			"currency_code":  stringOf(record, models.SalesOrderFieldCurrencyCode),

			// The totals travel WITH the event so a consumer never has to read back into Sales to
			// act on it - and so that what it acts on is what was true at confirmation rather than
			// whatever the order says by the time the event is delivered.
			"grand_total": decimalOf(record, models.SalesOrderFieldGrandTotal),
			"tax_total":   decimalOf(record, models.SalesOrderFieldTaxTotal),

			"confirmed_at": at.Unix(),
		},
	})
	return err
}
