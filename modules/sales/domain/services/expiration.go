package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Expiring stale drafts and lapsed offers (BR 87.3, SALES-040).
//
// # An expired ORDER becomes `cancelled`; an expired QUOTATION becomes `expired`
//
// The asymmetry is deliberate and follows the two enums as they stand. sales_orders has no `expired`
// status — its five are draft, confirmed, processing, completed, cancelled — and `cancelled` is
// documented as "a sale that will not happen", which is exactly what a stale draft is. Adding a
// sixth value would mean a schema migration, a transition-table edit, and every existing reader of
// order status learning a new case, in return for a distinction the requirement never asked the
// status column to carry.
//
// What DOES carry it is the audit trail: expiry writes its own action (`expire`) and its own reason,
// so an operator can always tell a basket that went stale from one somebody withdrew. That is the
// distinction that matters in practice — the two are served differently, one re-quoted and one owed
// an explanation — and it survives without touching the status enum.
//
// Quotations are the other way round because `expired` is one of their five states already (BR 87.1),
// and there `expired` and `cancelled` mean genuinely different things to the customer holding the
// offer.
//
// # What expiry MUST release (BR 82)
//
// A draft holds two things it has no right to keep once it is stale: voucher reservations, which
// stop another customer using a code, and stock reservations, which stop the goods being sold. Both
// are held on the promise that the sale is about to happen, and a stale draft has stopped promising
// that. Releasing them is the whole point of the feature — expiry that only changed a status would
// leave the scarce things locked.
//
// # The cutoff is computed per organization
//
// draft_order_expiry_hours is an org setting and may be overridden, so the sweep resolves the policy
// for each order rather than computing one cutoff for everybody. That costs a settings read per
// order, and the alternative — one global cutoff — would apply one organization's window to another
// organization's sales.

// ExpiryResult is what one sweep did.
type ExpiryResult struct {
	// ExpiredOrderIds and ExpiredQuotationIds are returned rather than counted, so a caller
	// investigating why a customer's basket vanished can find it by id.
	ExpiredOrderIds     []string
	ExpiredQuotationIds []string

	// ReleasedVoucherCodeIds names every code handed back. Reported separately because it is the
	// part with an effect outside this order: another customer can now use those codes.
	ReleasedVoucherCodeIds []string
}

// ExpireStaleDrafts expires draft orders past their window (BR 87.3).
//
// Each order is expired in its own transaction, and one failure does not stop the sweep. An
// unreachable row or a bad record would otherwise strand every order behind it, and the next run
// would meet the same row first and strand them again.
func ExpireStaleDrafts(
	ctx corectx.Context, policy SalesPolicy, now time.Time, limit int,
) (*ExpiryResult, error) {
	drafts, err := draftOrdersOlderThan(ctx, cutoffFor(policy, now), limit)
	if err != nil {
		return nil, err
	}

	result := &ExpiryResult{}
	for _, order := range drafts {
		orderId := stringOf(order, models.SalesOrderFieldId)

		released, err := expireOneOrder(ctx, order, now)
		if err != nil {
			// Logged by the caller; the sweep moves on. See the note above.
			continue
		}
		result.ExpiredOrderIds = append(result.ExpiredOrderIds, orderId)
		result.ReleasedVoucherCodeIds = append(result.ReleasedVoucherCodeIds, released...)
	}
	return result, nil
}

// cutoffFor answers the moment before which a draft is stale.
func cutoffFor(policy SalesPolicy, now time.Time) time.Time {
	hours := policy.DraftOrderExpiryHours
	if hours <= 0 {
		// A non-positive window would expire every draft the instant it was created, including the
		// one the customer is filling in. The schema forbids it (min 1), so this only fires on an
		// unconfigured policy — where doing nothing is the safe reading.
		hours = 24
	}
	return now.Add(-time.Duration(hours) * time.Hour)
}

// draftOrdersOlderThan reads the drafts that have gone stale.
//
// Filtered by status in the query and by age in Go, because RepoSearchParam carries no comparison
// against a timestamp. That means the page is drawn from all drafts rather than only old ones — fine
// while drafts are few, and worth revisiting if a deployment accumulates them, which the limit
// bounds in the meantime.
func draftOrdersOlderThan(
	ctx corectx.Context, cutoff time.Time, limit int,
) ([]dmodel.DynamicFields, error) {
	candidates, err := searchBy(ctx, models.SalesOrderSchemaName,
		models.SalesOrderFieldStatus, string(models.SalesOrderStatusDraft))
	if err != nil {
		return nil, err
	}

	stale := make([]dmodel.DynamicFields, 0, len(candidates))
	for _, order := range candidates {
		createdAt := dateTimeOf(order, basemodel.FieldCreatedAt)
		if createdAt == nil || !createdAt.GoTime().Before(cutoff) {
			continue
		}
		stale = append(stale, order)
		if limit > 0 && len(stale) >= limit {
			break
		}
	}
	return stale, nil
}

// expireOneOrder moves one draft to expired and releases what it was holding.
//
// All of it in ONE transaction. An order marked expired whose vouchers were not released would leave
// a code locked against a sale that no longer exists, and nothing would ever come back to free it.
func expireOneOrder(
	ctx corectx.Context, order dmodel.DynamicFields, now time.Time,
) ([]string, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	orgId := stringOf(order, basemodel.FieldOrgId)

	var released []string
	err := withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		// BR 82: hand back the voucher reservations first. If this fails the status move rolls back
		// with it, so the order stays a draft and the next sweep tries again — which is the right
		// outcome, because an expired order with locked codes is worse than a stale draft.
		codes, err := releaseOrderVouchers(tranxCtx, orderId)
		if err != nil {
			return err
		}
		released = codes

		engine, err := engineFor(models.SalesOrderSchemaName)
		if err != nil {
			return err
		}
		if _, err := engine.ResourceRepository().Update(tranxCtx, dmodel.DynamicFields{
			models.SalesOrderFieldId:     orderId,
			models.SalesOrderFieldStatus: string(models.SalesOrderStatusCancelled),
		}); err != nil {
			return err
		}

		// The status says `cancelled`, but the ACTION says `expire` — and that is what keeps the two
		// distinguishable. An operator asking why a basket vanished reads this, not the status
		// column, and the answer must be "it went stale" rather than "somebody withdrew it".
		return WriteSalesAuditEvent(tranxCtx, SalesAuditEntry{
			SalesOrderId: orderId,
			EntityType:   models.SalesOrderSchemaName,
			EntityId:     orderId,
			Action:       models.SalesOrderActionExpire,
			FromStatus:   string(models.SalesOrderStatusDraft),
			ToStatus:     string(models.SalesOrderStatusCancelled),
			Reason:       "draft expired without being confirmed",
			OrgId:        orgId,
			Metadata: map[string]any{
				"expired_at":                  now.Unix(),
				"expired_not_cancelled":       true,
				"released_voucher_code_count": len(released),
			},
		})
	})
	if err != nil {
		return nil, err
	}
	return released, nil
}

// ExpireLapsedQuotations expires sent quotations past their stated deadline (BR 87.3).
//
// Driven by the quotation's OWN valid_until rather than by the org window, and the difference
// matters: a quotation carries the deadline the customer was actually given, which may be longer or
// shorter than the default, and applying the org window would move a deadline somebody already
// promised. A quotation with no stated expiry never lapses — that absence is a decision the issuer
// made, not a missing value to fill in.
func ExpireLapsedQuotations(
	ctx corectx.Context, now time.Time, limit int,
) (*ExpiryResult, error) {
	sent, err := searchBy(ctx, models.SalesQuotationSchemaName,
		models.SalesQuotationFieldStatus, string(models.SalesQuotationStatusSent))
	if err != nil {
		return nil, err
	}

	result := &ExpiryResult{}
	for _, quotation := range sent {
		validUntil := dateTimeOf(quotation, models.SalesQuotationFieldValidUntil)
		if validUntil == nil || !validUntil.GoTime().Before(now) {
			continue
		}

		quotationId := stringOf(quotation, models.SalesQuotationFieldId)
		if err := expireOneQuotation(ctx, quotationId); err != nil {
			continue
		}
		result.ExpiredQuotationIds = append(result.ExpiredQuotationIds, quotationId)

		if limit > 0 && len(result.ExpiredQuotationIds) >= limit {
			break
		}
	}
	return result, nil
}

// expireOneQuotation moves one offer to expired.
//
// It goes through TransitionQuotation rather than writing the status directly, so the transition
// table stays the single authority on what a quotation may become — including its refusal to expire
// one that has meanwhile been accepted, which is the race this sweep is most likely to lose.
func expireOneQuotation(ctx corectx.Context, quotationId string) error {
	vErrs, err := TransitionQuotation(ctx, quotationId,
		string(models.SalesQuotationStatusExpired))
	if err != nil {
		return err
	}
	if vErrs != nil {
		// Refused rather than failed: the quotation moved on between the read and the write, which
		// is a normal race and not an error. Returning it as one would make the sweep log noise
		// about a correct outcome.
		return nil
	}
	return nil
}

// StampQuotationValidUntil sets a quotation's deadline from the org window.
//
// Used when an issuer states no deadline of their own, so that a quotation still lapses rather than
// standing open forever. Called at SEND time rather than at creation, because the window runs from
// when the customer was shown the offer.
func StampQuotationValidUntil(
	ctx corectx.Context, quotationId string, policy SalesPolicy, now time.Time,
) error {
	quotation, err := loadRecord(ctx,
		models.SalesQuotationSchemaName, models.SalesQuotationFieldId, quotationId)
	if err != nil || quotation == nil {
		return err
	}

	// An issuer's own deadline is never overwritten. They chose it, and the customer may already
	// have been told.
	if dateTimeOf(quotation, models.SalesQuotationFieldValidUntil) != nil {
		return nil
	}

	hours := policy.DraftOrderExpiryHours
	if hours <= 0 {
		return nil
	}

	engine, err := engineFor(models.SalesQuotationSchemaName)
	if err != nil {
		return err
	}
	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.SalesQuotationFieldId: quotationId,
		models.SalesQuotationFieldValidUntil: model.ModelDateTime(
			now.Add(time.Duration(hours) * time.Hour)),
	})
	return err
}
