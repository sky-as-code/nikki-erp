package services

import (
	"testing"

	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

func promotionOf(programType models.PromotionProgramType, status models.PromotionStatus) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		"id": "P1", "org_id": "O1", models.SalesPromotionFieldProgramType: string(programType),
		models.SalesPromotionFieldStatus: string(status), models.SalesPromotionFieldCurrencyId: "VND",
	}
}

func fieldsOf(vErrs *ft.ClientErrors) []string {
	fields := []string{}
	for _, item := range *vErrs {
		fields = append(fields, item.Field)
	}
	return fields
}

// BR-L02: an active promotion and its children are locked; draft and suspended are editable.
func TestPromotionEditLock(t *testing.T) {
	require.NotNil(t, assertPromotionEditable(promotionOf(models.PromotionTypeAutomatic, models.PromotionStatusActive)))
	require.Nil(t, assertPromotionEditable(promotionOf(models.PromotionTypeAutomatic, models.PromotionStatusDraft)))
	require.Nil(t, assertPromotionEditable(promotionOf(models.PromotionTypeAutomatic, models.PromotionStatusSuspended)))
}

// BR-V08: every type needs an apply rule; next_order_coupon also an issue rule, and no other type
// may have one. buy_x_get_y needs a qualification unit and a positive set size.
func TestActivationRules(t *testing.T) {
	apply := dmodel.DynamicFields{
		models.SalesPromotionRuleFieldPhase: "apply", models.SalesPromotionRuleFieldMinQuantity: "3"}
	issue := dmodel.DynamicFields{models.SalesPromotionRuleFieldPhase: "issue"}
	zeroSet := dmodel.DynamicFields{models.SalesPromotionRuleFieldPhase: "apply",
		models.SalesPromotionRuleFieldQualificationUomId: "U1", models.SalesPromotionRuleFieldMinQuantity: "0"}
	cases := []struct {
		name        string
		programType models.PromotionProgramType
		rules       []dmodel.DynamicFields
		violations  int
	}{
		{"no rule", models.PromotionTypeAutomatic, nil, 1},
		{"automatic apply", models.PromotionTypeAutomatic, []dmodel.DynamicFields{apply}, 0},
		{"automatic issue refused", models.PromotionTypeAutomatic, []dmodel.DynamicFields{apply, issue}, 1},
		{"next order both phases", models.PromotionTypeNextOrderCoupon, []dmodel.DynamicFields{apply, issue}, 0},
		{"next order without issue", models.PromotionTypeNextOrderCoupon, []dmodel.DynamicFields{apply}, 1},
		{"bxgy without unit", models.PromotionTypeBuyXGetY, []dmodel.DynamicFields{apply}, 1},
		{"bxgy zero set", models.PromotionTypeBuyXGetY, []dmodel.DynamicFields{zeroSet}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vErrs := ft.NewClientErrors()
			checkActivationRules(tc.programType, tc.rules, vErrs)
			require.Equal(t, tc.violations, vErrs.Count(), fieldsOf(vErrs))
		})
	}
}

func TestActivationNeedsCouponValidity(t *testing.T) {
	vErrs := ft.NewClientErrors()
	checkActivationWindow(promotionOf(models.PromotionTypeNextOrderCoupon, models.PromotionStatusDraft),
		models.PromotionTypeNextOrderCoupon, vErrs)
	require.Equal(t, []string{models.SalesPromotionFieldNextVoucherValidityDays}, fieldsOf(vErrs))
}

func TestRuleAndRewardWriteRules(t *testing.T) {
	vErrs := ft.NewClientErrors()
	checkRule(promotionOf(models.PromotionTypeCoupon, models.PromotionStatusDraft),
		dmodel.DynamicFields{models.SalesPromotionRuleFieldPhase: "issue"}, vErrs)
	require.Equal(t, 1, vErrs.Count())

	vErrs = ft.NewClientErrors()
	checkReward(promotionOf(models.PromotionTypeBuyXGetY, models.PromotionStatusDraft), dmodel.DynamicFields{
		models.SalesPromotionRewardFieldRewardType: "discount", models.SalesPromotionRewardFieldDiscountMethod: "percentage",
		models.SalesPromotionRewardFieldDiscountValue: "120"}, vErrs)
	require.ElementsMatch(t, []string{models.SalesPromotionRewardFieldRewardType,
		models.SalesPromotionRewardFieldDiscountValue}, fieldsOf(vErrs))

	require.Equal(t, []string{models.SalesPromotionRuleFieldQualificationUomId},
		clearRule(promotionOf(models.PromotionTypeAutomatic, models.PromotionStatusDraft), nil))
	require.Nil(t, clearRule(promotionOf(models.PromotionTypeBuyXGetY, models.PromotionStatusDraft), nil))
}

// BR §8.5: a field that does not belong to the reward type is cleared.
func TestClearRewardKeepsOnlyTheTypeFields(t *testing.T) {
	cleared := clearReward(nil, dmodel.DynamicFields{models.SalesPromotionRewardFieldRewardType: "free_product"})
	require.Contains(t, cleared, models.SalesPromotionRewardFieldDiscountValue)
	require.Contains(t, cleared, models.SalesPromotionRewardFieldMaxDiscountAmount)
	require.NotContains(t, cleared, models.SalesPromotionRewardFieldGiftProductVariantId)

	cleared = clearReward(nil, dmodel.DynamicFields{models.SalesPromotionRewardFieldRewardType: "free_shipping"})
	require.NotContains(t, cleared, models.SalesPromotionRewardFieldMaxDiscountAmount)
	require.Contains(t, cleared, models.SalesPromotionRewardFieldDiscountMethod)
}

// BR-V03: trim and uppercase only; the pattern decides the rest.
func TestVoucherCodeNormalizing(t *testing.T) {
	require.Equal(t, "SUMMER10", NormalizeVoucherCode("  summer10 "))
	require.True(t, voucherCodePattern.MatchString(NormalizeVoucherCode("ab-12cd")))
	require.False(t, voucherCodePattern.MatchString(NormalizeVoucherCode("SUM MER10")))
	require.False(t, voucherCodePattern.MatchString("ABC12"))
}

func TestChangedFields(t *testing.T) {
	stored := dmodel.DynamicFields{"program_type": "coupon", "currency_id": "VND"}
	require.Equal(t, []string{"program_type"}, changedFields(stored,
		dmodel.DynamicFields{"program_type": "automatic", "currency_id": "VND"}, "program_type", "currency_id"))
}
