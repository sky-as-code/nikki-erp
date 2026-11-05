package services

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// BR-V03, BR-V06: a generated code is the normalised prefix plus 8 secure random characters and
// always passes the code pattern.
func TestGenerateVoucherCode(t *testing.T) {
	code, err := GenerateVoucherCode(" sale- ")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(code, "SALE-"))
	require.Len(t, code, len("SALE-")+voucherRandomLength)
	require.True(t, voucherCodePattern.MatchString(code))

	other, err := GenerateVoucherCode("SALE-")
	require.NoError(t, err)
	require.NotEqual(t, code, other)

	_, err = GenerateVoucherCode(strings.Repeat("A", 25))
	require.Error(t, err)
	_, err = GenerateVoucherCode("SALE 1")
	require.Error(t, err)
}

func TestBatchRequestLimits(t *testing.T) {
	require.Nil(t, assertBatchRequest(1, ""))
	require.Nil(t, assertBatchRequest(voucherBatchMax, "KM-"))
	require.NotNil(t, assertBatchRequest(0, ""))
	require.NotNil(t, assertBatchRequest(voucherBatchMax+1, ""))
	require.NotNil(t, assertBatchRequest(5, "bad prefix"))
}

// BR-V01 and BR §9.2: automatic and buy X get Y take no voucher; a next-order coupon only from its
// source order.
func TestVoucherAllowedByProgramType(t *testing.T) {
	none := dmodel.DynamicFields{}
	sourced := dmodel.DynamicFields{models.SalesVoucherFieldSourceOrderId: "SO1"}
	require.NotNil(t, assertVoucherAllowed(promotionOf(models.PromotionTypeAutomatic, models.PromotionStatusActive), none))
	require.NotNil(t, assertVoucherAllowed(promotionOf(models.PromotionTypeBuyXGetY, models.PromotionStatusActive), none))
	require.Nil(t, assertVoucherAllowed(promotionOf(models.PromotionTypeDiscountCode, models.PromotionStatusActive), none))
	require.Nil(t, assertVoucherAllowed(promotionOf(models.PromotionTypeCoupon, models.PromotionStatusActive), none))
	noc := promotionOf(models.PromotionTypeNextOrderCoupon, models.PromotionStatusActive)
	require.NotNil(t, assertVoucherAllowed(noc, none))
	require.Nil(t, assertVoucherAllowed(noc, sourced))
}
