package services

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Voucher reservation and redemption.
//
// Two tills redeeming the last use of a usage_limit=1 voucher at the same instant must not both
// succeed, and the counter cannot be the safe point: this framework cannot express a conditional
// single-statement increment. DynamicResourceRepository.Update takes a field map, so no expression
// and no predicate beyond the primary key; MutateResultData.AffectedCount is hardcoded to 1, so a
// conditional update that matched nothing could not be detected; the versioned_model etag is
// compared in application code after a read, which is the same race one layer up; and
// FindOneForUpdate is typed for ent entities and has no callers.
//
// So the safe point is the unique index on (voucher_code_id, sales_order_id), which Postgres
// enforces regardless of what any process believes. The reservation row is written FIRST; if it
// lands, this order holds the use. usage_count is then maintained as a derived cache of the
// redemption rows.
//
// The consequence: a usage_limit of N can be over-reserved by concurrent requests, since N separate
// orders each write a row that collides with nothing. The index prevents double-spend by ONE order,
// not over-issue across many; closing that needs SELECT ... FOR UPDATE or a real conditional UPDATE,
// neither reachable without extending the framework. For usage_limit = 1 the two coincide and the
// index does close it, provided the reservation is written before the discount is honoured.

// ReserveVoucher holds one use of a code for one order.
//
// Called when a voucher is applied to a draft. The reservation is what stops a second customer
// taking the last use of a code already sitting in someone's basket; a usage counter alone could
// not, because a draft has not incremented anything yet.
//
// Returns ClientErrors when the code cannot be reserved: the caller must surface those as a 400
// with the machine-readable reason, never as a server error.
func ReserveVoucher(
	ctx corectx.Context, voucherCodeId, salesOrderId, orgId string, nowUnix int64,
) (*models.SalesVoucherRedemption, *ft.ClientErrors, error) {
	codeRecord, err := loadRecord(ctx,
		models.SalesVoucherCodeSchemaName, models.SalesVoucherCodeFieldId, voucherCodeId)
	if err != nil {
		return nil, nil, err
	}
	if codeRecord == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonCodeNotFound,
			"no voucher exists with that code"))
		return nil, vErrs, nil
	}

	code := models.NewSalesVoucherCodeFrom(codeRecord)
	if vErrs := assertCodeUsable(code, nowUnix); vErrs != nil {
		return nil, vErrs, nil
	}

	redemption, err := insertReservation(ctx, voucherCodeId, salesOrderId, orgId)
	if err != nil {
		// A collision on (voucher_code_id, sales_order_id) is the database refusing a second
		// reservation of this code by this order. A client error, not a fault: the caller applied the
		// same voucher twice.
		if isUniqueViolation(err) {
			vErrs := ft.NewClientErrors()
			vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonAlreadyApplied,
				"this voucher is already applied to this order"))
			return nil, vErrs, nil
		}
		return nil, nil, err
	}
	return redemption, nil, nil
}

// The machine-readable refusal reasons. Codes rather than sentences so a till can branch on them
// and a translation layer can render them; the accompanying message is for a log or a developer.
const (
	ReasonCodeNotFound      = "sales_voucher.not_found"
	ReasonAlreadyApplied    = "sales_voucher.already_applied"
	ReasonExpired           = "sales_voucher.expired"
	ReasonNotYetValid       = "sales_voucher.not_yet_valid"
	ReasonDisabled          = "sales_voucher.disabled"
	ReasonUsageExhausted    = "sales_voucher.usage_exhausted"
	ReasonArchived          = "sales_voucher.archived"
	ReasonNotReserved       = "sales_voucher.not_reserved"
	ReasonNotRedeemed       = "sales_voucher.not_redeemed"
	ReasonIllegalTransition = "sales_voucher.illegal_transition"
)

// isUniqueViolation reports whether an insert failed because it collided with a unique index.
//
// Matched on the message because the repository returns a wrapped driver error rather than a typed
// one; the same three spellings jobscheduler matches on cover every driver this deployment uses. A
// missed match degrades a 400 into a 500, and the row is not written either way, so the invariant
// holds regardless.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "sqlstate 23505")
}

// assertCodeUsable applies the gates that live on the code record itself.
//
// Each returns its own reason rather than a shared "unusable", because an expired code and an
// exhausted one call for different responses at the till. Gate order is expired-before-exhausted
// deliberately: a code that is both is more usefully described as expired, since that is the
// condition the customer can see on the voucher itself.
func assertCodeUsable(code *models.SalesVoucherCode, nowUnix int64) *ft.ClientErrors {
	vErrs := ft.NewClientErrors()

	if archived := code.GetIsArchived(); archived != nil && *archived {
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonArchived,
			"this voucher has been withdrawn"))
		return vErrs
	}

	if from := code.GetValidFrom(); from != nil && nowUnix < from.GoTime().Unix() {
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonNotYetValid,
			"this voucher is not valid yet"))
		return vErrs
	}
	if until := code.GetValidUntil(); until != nil && nowUnix >= until.GoTime().Unix() {
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonExpired,
			"this voucher has expired"))
		return vErrs
	}

	status := code.GetStatus()
	switch {
	case status == nil:
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonDisabled,
			"this voucher has no status and cannot be used"))
		return vErrs
	case *status == string(models.VoucherCodeStatusDisabled):
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonDisabled,
			"this voucher has been disabled"))
		return vErrs
	case *status == string(models.VoucherCodeStatusExhausted):
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonUsageExhausted,
			"this voucher has no uses left"))
		return vErrs
	}

	if !code.HasUsesRemaining() {
		vErrs.Append(*ft.NewBusinessViolation("voucher_code", ReasonUsageExhausted,
			"this voucher has no uses left"))
		return vErrs
	}
	return nil
}

// insertReservation writes the row whose unique index is the actual concurrency control.
func insertReservation(
	ctx corectx.Context, voucherCodeId, salesOrderId, orgId string,
) (*models.SalesVoucherRedemption, error) {
	engine, err := engineFor(models.SalesVoucherRedemptionSchemaName)
	if err != nil {
		return nil, err
	}

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "ReserveVoucher")
	}

	fields := dmodel.DynamicFields{
		models.SalesVoucherRedemptionFieldId:            string(*id),
		models.SalesVoucherRedemptionFieldVoucherCodeId: voucherCodeId,
		models.SalesVoucherRedemptionFieldSalesOrderId:  salesOrderId,
		models.SalesVoucherRedemptionFieldStatus:        string(models.VoucherRedemptionStatusReserved),
		models.SalesVoucherRedemptionFieldReservedAt:    model.ModelDateTime(time.Now().UTC()),
	}
	if orgId != "" {
		fields[basemodel.FieldOrgId] = orgId
	}

	// Through the repository rather than the resource service: the redemption ledger is read-only to
	// clients, and must stay so while the system writes its own rows.
	if _, err := engine.ResourceRepository().Insert(ctx, fields); err != nil {
		return nil, err
	}
	return models.NewSalesVoucherRedemptionFrom(fields), nil
}

// SettleRedemption moves a redemption to a new status and keeps usage_count in step.
//
// The four settlements share one function because they differ only in the target status and which
// timestamp they stamp; four copies would be four places for the counter maintenance to drift. The
// transition is checked against the table in order_states.go, so an illegal move is refused rather
// than silently written: a released reservation being redeemed would give away a discount the order
// never held.
func SettleRedemption(
	ctx corectx.Context, redemptionId, toStatus string,
) (*ft.ClientErrors, error) {
	record, err := loadRecord(ctx,
		models.SalesVoucherRedemptionSchemaName, models.SalesVoucherRedemptionFieldId, redemptionId)
	if err != nil {
		return nil, err
	}
	if record == nil {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("redemption", ReasonNotReserved,
			"no redemption exists with that id"))
		return vErrs, nil
	}

	redemption := models.NewSalesVoucherRedemptionFrom(record)
	from := ""
	if status := redemption.GetStatus(); status != nil {
		from = *status
	}

	if !CanTransitionVoucherRedemption(from, toStatus) {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("redemption", ReasonIllegalTransition,
			"a redemption cannot move from '"+from+"' to '"+toStatus+"'"))
		return vErrs, nil
	}
	if from == toStatus {
		// Already there. Allowed as a no-op so a retry is not an error, and returning early is what
		// keeps it one: falling through would stamp the timestamp twice and adjust the counter for a
		// move that did not happen.
		return nil, nil
	}

	update := dmodel.DynamicFields{
		models.SalesVoucherRedemptionFieldId:     redemptionId,
		models.SalesVoucherRedemptionFieldStatus: toStatus,
	}
	switch toStatus {
	case string(models.VoucherRedemptionStatusRedeemed):
		update[models.SalesVoucherRedemptionFieldRedeemedAt] = model.ModelDateTime(time.Now().UTC())
	case string(models.VoucherRedemptionStatusReleased):
		update[models.SalesVoucherRedemptionFieldReleasedAt] = model.ModelDateTime(time.Now().UTC())
	case string(models.VoucherRedemptionStatusReversed):
		update[models.SalesVoucherRedemptionFieldReversedAt] = model.ModelDateTime(time.Now().UTC())
	}

	engine, err := engineFor(models.SalesVoucherRedemptionSchemaName)
	if err != nil {
		return nil, err
	}
	if _, err := engine.ResourceRepository().Update(ctx, update); err != nil {
		return nil, err
	}

	// The counter follows the ledger. A move that stops holding a use gives one back; nothing else
	// changes it, because reserving already took it.
	if wasHolding(from) && !wasHolding(toStatus) {
		codeId := ""
		if id := redemption.GetVoucherCodeId(); id != nil {
			codeId = string(*id)
		}
		if err := refreshUsageCount(ctx, codeId); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func wasHolding(status string) bool {
	return status == string(models.VoucherRedemptionStatusReserved) ||
		status == string(models.VoucherRedemptionStatusRedeemed)
}

// refreshUsageCount recomputes a code's usage_count from its redemption rows.
//
// Recounted rather than decremented: a decrement assumes the stored value was right to begin with
// and compounds the error if it was not, whereas a recount is self-correcting and keeps the counter
// honestly a cache of the ledger. It also updates `status`, so a code that has run out reads as
// exhausted and one that has a use returned to it becomes active again.
func refreshUsageCount(ctx corectx.Context, voucherCodeId string) error {
	if voucherCodeId == "" {
		return nil
	}

	held, err := countHeldRedemptions(ctx, voucherCodeId)
	if err != nil {
		return err
	}

	codeRecord, err := loadRecord(ctx,
		models.SalesVoucherCodeSchemaName, models.SalesVoucherCodeFieldId, voucherCodeId)
	if err != nil || codeRecord == nil {
		return err
	}
	code := models.NewSalesVoucherCodeFrom(codeRecord)

	update := dmodel.DynamicFields{
		models.SalesVoucherCodeFieldId:         voucherCodeId,
		models.SalesVoucherCodeFieldUsageCount: held,
	}

	// Only ever move between active and exhausted. A disabled code stays disabled: that is an
	// operator's decision, and a return restoring a use must not quietly switch a campaign back on.
	if status := code.GetStatus(); status != nil &&
		*status != string(models.VoucherCodeStatusDisabled) {
		limit := code.GetUsageLimit()
		switch {
		case limit == nil:
			update[models.SalesVoucherCodeFieldStatus] = string(models.VoucherCodeStatusActive)
		case held >= *limit:
			update[models.SalesVoucherCodeFieldStatus] = string(models.VoucherCodeStatusExhausted)
		default:
			update[models.SalesVoucherCodeFieldStatus] = string(models.VoucherCodeStatusActive)
		}
	}

	engine, err := engineFor(models.SalesVoucherCodeSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, update)
	return err
}

// countHeldRedemptions counts the rows currently holding a use of this code.
func countHeldRedemptions(ctx corectx.Context, voucherCodeId string) (int32, error) {
	engine, err := engineFor(models.SalesVoucherRedemptionSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().
		NewCondition(models.SalesVoucherRedemptionFieldVoucherCodeId, dmodel.Equals, voucherCodeId))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  model.MODEL_RULE_PAGE_MAX_SIZE,
	})
	if err != nil {
		return 0, errors.Wrap(err, "countHeldRedemptions")
	}
	if found == nil || !found.HasData {
		return 0, nil
	}

	// Counted in Go rather than filtered in the query because the status set is small and the
	// predicate is the same one wasHolding uses; two spellings of the same rule would eventually
	// disagree about whether a reservation counts.
	var held int32
	for _, item := range found.Data.Items {
		redemption := models.NewSalesVoucherRedemptionFrom(item)
		if redemption.HoldsAUse() {
			held++
		}
	}
	return held, nil
}
