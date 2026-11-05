package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services/pricing/promotion"
)

// The pure parts of the order-side promotion flow: voucher availability, why an attached voucher is
// rejected, the published reason codes, the requote compare and the untaxed lines of the second tax
// pass. The repository-bound paths are exercised by the API suites.

var flowNow = time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)

func voucherRow(fields dmodel.DynamicFields) dmodel.DynamicFields {
	row := dmodel.DynamicFields{
		basemodel.FieldId:              "v1",
		models.SalesVoucherFieldStatus: string(models.VoucherStatusActive),
		models.SalesVoucherFieldCode:   "SUMMER-01",
	}
	for key, value := range fields {
		row[key] = value
	}
	return row
}

func TestVoucherUsableAt(t *testing.T) {
	at := func(offset time.Duration) model.ModelDateTime { return model.ModelDateTime(flowNow.Add(offset)) }
	cases := []struct {
		name   string
		fields dmodel.DynamicFields
		usable bool
	}{
		{name: "active, open window", usable: true},
		{name: "archived", fields: dmodel.DynamicFields{basemodel.FieldIsArchived: true}},
		{name: "revoked", fields: dmodel.DynamicFields{models.SalesVoucherFieldStatus: string(models.VoucherStatusRevoked)}},
		{name: "not started", fields: dmodel.DynamicFields{models.SalesVoucherFieldValidFrom: at(time.Hour)}},
		{name: "AC-23 at exactly expires_at", fields: dmodel.DynamicFields{models.SalesVoucherFieldExpiresAt: at(0)}},
		{name: "AC-23 at exactly valid_from", usable: true,
			fields: dmodel.DynamicFields{models.SalesVoucherFieldValidFrom: at(0)}},
		{name: "single use already used", fields: dmodel.DynamicFields{
			models.SalesVoucherFieldUsageLimit: int32(1), models.SalesVoucherFieldUsedCount: int32(1)}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.usable, VoucherUsableAt(voucherRow(testCase.fields), flowNow))
		})
	}
	require.False(t, VoucherUsableAt(nil, flowNow))
}

func TestVoucherRejectReason(t *testing.T) {
	pc := &promotionContext{orderId: "order-1", customerId: "cus-1", now: flowNow}
	coupon := dmodel.DynamicFields{models.SalesPromotionFieldProgramType: string(models.PromotionTypeCoupon)}
	automatic := dmodel.DynamicFields{models.SalesPromotionFieldProgramType: string(models.PromotionTypeAutomatic)}
	cases := []struct {
		name      string
		promotion dmodel.DynamicFields
		voucher   dmodel.DynamicFields
		live      bool
		scoped    bool
		context   bool
		reason    string
	}{
		{name: "usable", promotion: coupon, voucher: voucherRow(nil), live: true, scoped: true, context: true},
		{name: "AC-05 no channel or point", promotion: coupon, voucher: voucherRow(nil), live: true, scoped: true,
			reason: PromotionReasonContextRequired},
		{name: "AC-22 another customer", promotion: coupon, live: true, scoped: true, context: true,
			voucher: voucherRow(dmodel.DynamicFields{models.SalesVoucherFieldCustomerId: "cus-2"}),
			reason:  PromotionReasonCustomerMismatch},
		{name: "AC-22 out of scope", promotion: coupon, voucher: voucherRow(nil), live: true, context: true,
			reason: PromotionReasonScopeMismatch},
		{name: "promotion suspended or expired", promotion: coupon, voucher: voucherRow(nil), scoped: true,
			context: true, reason: PromotionReasonNotAvailable},
		{name: "BR-V09 own source order", promotion: coupon, live: true, scoped: true, context: true,
			voucher: voucherRow(dmodel.DynamicFields{models.SalesVoucherFieldSourceOrderId: "order-1"}),
			reason:  PromotionReasonNotAvailable},
		{name: "BR-V01 automatic takes no voucher", promotion: automatic, voucher: voucherRow(nil), live: true,
			scoped: true, context: true, reason: PromotionReasonNotAvailable},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.reason, pc.voucherRejectReason(
				testCase.promotion, testCase.voucher, testCase.live, testCase.scoped, testCase.context))
		})
	}
}

func TestPromotionLiveAt(t *testing.T) {
	window := dmodel.DynamicFields{
		models.SalesPromotionFieldStartsAt: model.ModelDateTime(flowNow),
		models.SalesPromotionFieldEndsAt:   model.ModelDateTime(flowNow.Add(time.Hour)),
	}
	require.True(t, promotionLiveAt(window, flowNow), "AC-23 valid at starts_at")
	require.False(t, promotionLiveAt(window, flowNow.Add(time.Hour)), "AC-23 expired at ends_at")
	exhausted := dmodel.DynamicFields{
		models.SalesPromotionFieldUsageLimit: int32(3), models.SalesPromotionFieldUsedCount: int32(3),
	}
	require.False(t, promotionLiveAt(exhausted, flowNow))
}

func TestPublishedReasonsAreTheTranslatedCodes(t *testing.T) {
	translated := map[string]bool{
		"context_required": true, "scope_mismatch": true, "not_available": true, "condition_not_met": true,
		"pricelist_mismatch": true, "no_shipping_line": true, "order_discount_not_best": true,
		"zero_benefit": true, "customer_required": true,
	}
	for engineReason := range engineReasons {
		require.True(t, translated[publishedReason(engineReason)], engineReason)
	}
	require.Equal(t, "", publishedReason(""))
}

// D4 / BR-T02: a different fresh reference refuses with the fresh quote under vars.quote.
func TestAssertQuoteAccepted(t *testing.T) {
	quote := &PromotionQuoteResult{Reference: "fresh", pc: &promotionContext{}, giftLineIds: map[string]string{}}
	require.Nil(t, assertQuoteAccepted("", quote), "never quoted: nothing to compare")
	require.Nil(t, assertQuoteAccepted("fresh", quote))
	vErrs := assertQuoteAccepted("stale", quote)
	require.NotNil(t, vErrs)
	item := (*vErrs)[0]
	require.Equal(t, "sales:"+PromotionErrRequoteRequired, item.Key)
	view, ok := item.Vars["quote"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "fresh", view["calculation_reference"])
}

// The exclusive tax pass sees each line's untaxed value after the promotions.
func TestUntaxedAfter(t *testing.T) {
	quote := &PromotionQuoteResult{
		engine: promotion.Input{Lines: []promotion.Line{
			{Key: "a", UntaxedAmount: decimal.RequireFromString("100000")},
			{Key: "b", UntaxedAmount: decimal.RequireFromString("50000")},
		}},
		outcome: promotion.Result{Lines: []promotion.LineOutcome{
			{Key: "a", RemainingUntaxedAmount: decimal.RequireFromString("90000")},
		}},
	}
	lines := []pricing.LineResult{
		{Key: "a", Quantity: decimal.RequireFromString("2")},
		{Key: "b", Quantity: decimal.RequireFromString("1")},
	}
	described := untaxedAfter(lines, quote)
	require.True(t, described[0].NetAmount.Equal(decimal.RequireFromString("90000")))
	require.True(t, described[0].EffectiveUnitPrice.Equal(decimal.RequireFromString("45000")))
	require.True(t, described[1].NetAmount.Equal(decimal.RequireFromString("50000")))
}

func TestQualificationUomsOnlyFromBuyXGetY(t *testing.T) {
	programs := []promotion.Program{
		{ProgramType: promotion.ProgramAutomatic, Rules: []promotion.Rule{{QualificationUomId: "uom-x"}}},
		{ProgramType: promotion.ProgramBuyXGetY, Rules: []promotion.Rule{{QualificationUomId: "uom-bottle"}}},
	}
	require.Equal(t, []string{"uom-bottle"}, qualificationUomsOf(programs))
}
