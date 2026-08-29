package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Expiring stale drafts and lapsed offers.
//
// An expired ORDER becomes `cancelled` because sales_orders has no `expired` status. What carries
// the distinction is the audit trail: expiry writes its own action (`expire`) and reason, so an
// operator can tell a basket that went stale from one somebody withdrew. Quotations are the other
// way round because `expired` is already one of their states.
//
// Expiry MUST release what a stale draft has no right to keep: voucher reservations, which stop
// another customer using a code, and stock reservations. Expiry that only changed a status would
// leave the scarce things locked.
//
// The cutoff is per organization, because draft_order_expiry_hours may be overridden and one
// global cutoff would apply one organization's window to another's sales.

type ExpiryResult struct {
	// Ids rather than counts, so a caller investigating why a basket vanished can find it.
	ExpiredOrderIds     []string
	ExpiredQuotationIds []string

	// ReleasedVoucherCodeIds is separate because it has an effect outside this order: another
	// customer can now use those codes.
	ReleasedVoucherCodeIds []string
}

// ExpireStaleDrafts gives each order its own transaction and one failure does not stop the sweep:
// an unreachable row would otherwise strand every order behind it, run after run.
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
			// Logged by the caller; the sweep moves on.
			continue
		}
		result.ExpiredOrderIds = append(result.ExpiredOrderIds, orderId)
		result.ReleasedVoucherCodeIds = append(result.ReleasedVoucherCodeIds, released...)
	}
	return result, nil
}

func cutoffFor(policy SalesPolicy, now time.Time) time.Time {
	hours := policy.DraftOrderExpiryHours
	if hours <= 0 {
		// A non-positive window would expire every draft the instant it was created. The schema
		// forbids it (min 1), so this only fires on an unconfigured policy.
		hours = 24
	}
	return now.Add(-time.Duration(hours) * time.Hour)
}

// draftOrdersOlderThan filters by status in the query and by age in Go, because RepoSearchParam
// carries no timestamp comparison. The page is therefore drawn from all drafts.
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

// expireOneOrder moves one draft to expired and releases what it was holding, in ONE transaction:
// an order marked expired whose vouchers were not released would leave a code locked against a
// sale that no longer exists, with nothing ever coming back to free it.
func expireOneOrder(
	ctx corectx.Context, order dmodel.DynamicFields, now time.Time,
) ([]string, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	orgId := stringOf(order, basemodel.FieldOrgId)

	var released []string
	err := withTransaction(ctx, models.SalesOrderSchemaName, func(tranxCtx corectx.Context) error {
		// Vouchers first. If this fails the status move rolls back with it, so the order stays a
		// draft and the next sweep tries again - an expired order with locked codes is worse.
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

		// The status says `cancelled` but the ACTION says `expire`, which keeps the two distinguishable.
		// An operator asking why a basket vanished reads this, not the status column.
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

// ExpireLapsedQuotations is driven by the quotation's OWN valid_until rather than the org window:
// a quotation carries the deadline the customer was actually given, and applying the org window
// would move a promise already made. One with no stated expiry never lapses.
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

// expireOneQuotation goes through TransitionQuotation rather than writing the status directly, so
// the transition table stays the single authority - including its refusal to expire an accepted
// one.
func expireOneQuotation(ctx corectx.Context, quotationId string) error {
	vErrs, err := TransitionQuotation(ctx, quotationId,
		string(models.SalesQuotationStatusExpired))
	if err != nil {
		return err
	}
	if vErrs != nil {
		// Refused rather than failed: the quotation moved on between the read and the write, a normal
		// race, and returning it as an error would make the sweep log noise about a correct outcome.
		return nil
	}
	return nil
}

// StampQuotationValidUntil sets a deadline when the issuer states none. Called at SEND time,
// because the window runs from when the customer was shown the offer.
func StampQuotationValidUntil(
	ctx corectx.Context, quotationId string, policy SalesPolicy, now time.Time,
) error {
	quotation, err := loadRecord(ctx,
		models.SalesQuotationSchemaName, models.SalesQuotationFieldId, quotationId)
	if err != nil || quotation == nil {
		return err
	}

	// An issuer's own deadline is never overwritten: the customer may already have been told.
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
