package services

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// Reversing the promotions of a cancelled order (BR-T07, BR-V12, plan §3.3). Cancel runs under the
// order lock, so the confirmed status of an application is the once-only guard: a retried cancel
// finds nothing confirmed and gives no use back twice (AC-35).

// promotionReversalReason is recorded on every application a cancel reverses.
const promotionReversalReason = "order_cancelled"

// assertSourceCouponsUnused refuses to cancel an order whose next-order coupon was already redeemed
// on another confirmed order (BR-V12, AC-34): that order must be handled first. Read-only, so it runs
// before cancel releases stock or raises a refund.
func assertSourceCouponsUnused(ctx corectx.Context, orderId string) (*ft.ClientErrors, error) {
	issuances, err := searchBy(ctx, models.SalesVoucherIssuanceSchemaName,
		models.SalesVoucherIssuanceFieldSourceOrderId, orderId)
	if err != nil || len(issuances) == 0 {
		return nil, err
	}
	used, err := searchRows(ctx, models.SalesPromotionApplicationSchemaName, false,
		inCondition(models.SalesPromotionApplicationFieldVoucherId,
			idsOf(issuances, models.SalesVoucherIssuanceFieldVoucherId)),
		equalsCondition(models.SalesPromotionApplicationFieldStatus, string(models.PromotionApplicationConfirmed)))
	if err != nil || len(used) == 0 {
		return nil, err
	}
	usingOrderId := stringOf(used[0], models.SalesPromotionApplicationFieldOrderId)
	return promotionRefusal("", PromotionErrSourceOrderVoucherUsed,
		"a coupon this order issued was used on order '"+usingOrderId+"'; handle that order first",
		map[string]any{"order_id": usingOrderId}), nil
}

// reversePromotions runs inside the cancel transaction: it reverses the confirmed applications,
// gives each use back once, and revokes the coupons the order issued (never re-activated, BR-T07).
func reversePromotions(ctx corectx.Context, orderId string, at time.Time) error {
	applications, err := searchRows(ctx, models.SalesPromotionApplicationSchemaName, false,
		equalsCondition(models.SalesPromotionApplicationFieldOrderId, orderId),
		equalsCondition(models.SalesPromotionApplicationFieldStatus, string(models.PromotionApplicationConfirmed)))
	if err != nil {
		return err
	}
	for _, application := range applications {
		if err := reverseApplication(ctx, application, at); err != nil {
			return err
		}
	}
	return revokeIssuedCoupons(ctx, orderId)
}

func reverseApplication(ctx corectx.Context, application dmodel.DynamicFields, at time.Time) error {
	repo, err := repoFor(models.SalesPromotionApplicationSchemaName)
	if err != nil {
		return err
	}
	// is_confirmed goes to NULL, not false: the partial unique on (order, promotion) counts only the
	// confirmed row, and a reversed one must not block a later confirmation.
	_, err = repo.Update(ctx, dmodel.DynamicFields{
		basemodel.FieldId: stringOf(application, basemodel.FieldId),
		models.SalesPromotionApplicationFieldStatus:         string(models.PromotionApplicationReversed),
		models.SalesPromotionApplicationFieldIsConfirmed:    nil,
		models.SalesPromotionApplicationFieldReversedAt:     model.ModelDateTime(at),
		models.SalesPromotionApplicationFieldReversalReason: promotionReversalReason,
	})
	if err != nil {
		return err
	}
	if err := releaseOne(ctx, models.SalesPromotionSchemaName,
		stringOf(application, models.SalesPromotionApplicationFieldPromotionId)); err != nil {
		return err
	}
	if voucherId := stringOf(application, models.SalesPromotionApplicationFieldVoucherId); voucherId != "" {
		return releaseOne(ctx, models.SalesVoucherSchemaName, voucherId)
	}
	return nil
}

func releaseOne(ctx corectx.Context, schemaName, id string) error {
	counter, err := usageCounterOf(schemaName)
	if err != nil {
		return err
	}
	_, err = counter.ReleaseUsage(ctx, id)
	return err
}

func revokeIssuedCoupons(ctx corectx.Context, orderId string) error {
	issuances, err := searchBy(ctx, models.SalesVoucherIssuanceSchemaName,
		models.SalesVoucherIssuanceFieldSourceOrderId, orderId)
	if err != nil {
		return err
	}
	for _, voucherId := range idsOf(issuances, models.SalesVoucherIssuanceFieldVoucherId) {
		vErrs, err := RevokeVoucherBySystem(ctx, voucherId, models.VoucherRevocationSourceOrderCancelled)
		if err != nil {
			return err
		}
		if vErrs != nil {
			return &promotionCommitRefusal{cErrs: *vErrs}
		}
	}
	return nil
}
