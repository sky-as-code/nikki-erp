package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// issueNextOrderCoupons issues at most one single-use coupon per next_order_coupon promotion whose
// issue phase the confirmed order meets (BR-V09..V11). It consumes no usage, applies nothing on this
// order, and is skipped when the order has no customer (the quote warned) or no positive validity
// window is left.
func issueNextOrderCoupons(
	ctx corectx.Context, order dmodel.DynamicFields, quote *PromotionQuoteResult, at time.Time,
) ([]IssuedVoucher, error) {
	issued := []IssuedVoucher{}
	for _, decision := range quote.issues {
		if !decision.Qualified {
			continue
		}
		candidate := quote.pc.byPromotionId(decision.PromotionId)
		if candidate == nil {
			continue
		}
		voucher, err := issueOneCoupon(ctx, order, candidate, decision, at)
		if err != nil {
			return nil, err
		}
		if voucher != nil {
			issued = append(issued, *voucher)
		}
	}
	return issued, nil
}

func issueOneCoupon(
	ctx corectx.Context, order dmodel.DynamicFields, candidate *promotionCandidate,
	decision promotion.IssueDecision, at time.Time,
) (*IssuedVoucher, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)
	existing, err := searchRows(ctx, models.SalesVoucherIssuanceSchemaName, false,
		equalsCondition(models.SalesVoucherIssuanceFieldSourceOrderId, orderId),
		equalsCondition(models.SalesVoucherIssuanceFieldPromotionId, decision.PromotionId))
	if err != nil || len(existing) > 0 {
		return nil, err
	}
	validFrom, expiresAt, ok, err := promotion.IssueValidity(
		at, validityDaysOf(candidate.record), endsAtOf(candidate.record))
	if err != nil || !ok {
		return nil, err
	}
	params := IssueCouponParams{
		OrgId: stringOf(order, basemodel.FieldOrgId), PromotionId: decision.PromotionId,
		SourceOrderId: orderId, CustomerId: stringOf(order, models.SalesOrderFieldSoldToPartyId),
		ValidFrom: model.ModelDateTime(validFrom),
	}
	if expiresAt != nil {
		end := model.ModelDateTime(*expiresAt)
		params.ExpiresAt = &end
	}
	voucher, vErrs, err := IssueNextOrderCoupon(ctx, params)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return nil, &promotionCommitRefusal{cErrs: *vErrs}
	}
	return voucher, insertIssuance(ctx, order, decision, voucher.Id, at)
}

func validityDaysOf(record dmodel.DynamicFields) *int32 {
	days := int32Of(record, models.SalesPromotionFieldNextVoucherValidityDays)
	if days <= 0 {
		return nil
	}
	return &days
}

func endsAtOf(record dmodel.DynamicFields) *time.Time {
	endsAt := dateTimeOf(record, models.SalesPromotionFieldEndsAt)
	if endsAt == nil {
		return nil
	}
	end := endsAt.GoTime()
	return &end
}

func insertIssuance(
	ctx corectx.Context, order dmodel.DynamicFields, decision promotion.IssueDecision, voucherId string, at time.Time,
) error {
	repo, err := repoFor(models.SalesVoucherIssuanceSchemaName)
	if err != nil {
		return err
	}
	id, err := model.NewId()
	if err != nil {
		return err
	}
	_, err = repo.Insert(ctx, dmodel.DynamicFields{
		basemodel.FieldId:                                 string(*id),
		basemodel.FieldOrgId:                              stringOf(order, basemodel.FieldOrgId),
		models.SalesVoucherIssuanceFieldSourceOrderId:     stringOf(order, models.SalesOrderFieldId),
		models.SalesVoucherIssuanceFieldPromotionId:       decision.PromotionId,
		models.SalesVoucherIssuanceFieldVoucherId:         voucherId,
		models.SalesVoucherIssuanceFieldIssueRuleSnapshot: issueSnapshot(decision),
		models.SalesVoucherIssuanceFieldIssuedAt:          model.ModelDateTime(at),
	})
	return err
}

func issueSnapshot(decision promotion.IssueDecision) map[string]any {
	snapshot := qualificationSnapshot(decision.Qualification)
	snapshot["config_version"] = decision.ConfigVersion
	return snapshot
}

// loadConfirmedPromotions answers what an earlier confirm committed, for a replay (BR-T05).
func loadConfirmedPromotions(ctx corectx.Context, orderId string) (*ConfirmedPromotions, error) {
	confirmed := &ConfirmedPromotions{ApplicationIds: []string{}, IssuedVouchers: []IssuedVoucher{}}
	applications, err := searchRows(ctx, models.SalesPromotionApplicationSchemaName, false,
		equalsCondition(models.SalesPromotionApplicationFieldOrderId, orderId),
		equalsCondition(models.SalesPromotionApplicationFieldStatus, string(models.PromotionApplicationConfirmed)))
	if err != nil {
		return nil, err
	}
	confirmed.ApplicationIds = idsOf(applications, basemodel.FieldId)
	issuances, err := searchBy(ctx, models.SalesVoucherIssuanceSchemaName,
		models.SalesVoucherIssuanceFieldSourceOrderId, orderId)
	if err != nil || len(issuances) == 0 {
		return confirmed, err
	}
	vouchers, err := searchRows(ctx, models.SalesVoucherSchemaName, true,
		inCondition(basemodel.FieldId, idsOf(issuances, models.SalesVoucherIssuanceFieldVoucherId)))
	if err != nil {
		return nil, err
	}
	for _, voucher := range vouchers {
		confirmed.IssuedVouchers = append(confirmed.IssuedVouchers, IssuedVoucher{
			Id: stringOf(voucher, basemodel.FieldId), Code: stringOf(voucher, models.SalesVoucherFieldCode),
		})
	}
	return confirmed, nil
}
