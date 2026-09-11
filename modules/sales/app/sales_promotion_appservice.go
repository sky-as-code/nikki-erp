package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/promotion"
)

// The promotion and voucher application services. No custom actions: a discount is defined here
// and applied by the order's apply_voucher, which lives on the order.

func NewSalesPromotionProgramApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionProgramApplicationService {
	return &SalesPromotionProgramApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionProgramApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPromotionConditionGroupApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionConditionGroupApplicationService {
	return &SalesPromotionConditionGroupApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionConditionGroupApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPromotionConditionApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionConditionApplicationService {
	return &SalesPromotionConditionApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionConditionApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPromotionConditionTargetApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionConditionTargetApplicationService {
	return &SalesPromotionConditionTargetApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionConditionTargetApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPromotionRewardApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionRewardApplicationService {
	return &SalesPromotionRewardApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionRewardApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesPromotionCompatibilityApplicationService(base composable.CrudApplicationService) itPromotion.SalesPromotionCompatibilityApplicationService {
	return &SalesPromotionCompatibilityApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionCompatibilityApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesVoucherCodeApplicationService(base composable.CrudApplicationService) itPromotion.SalesVoucherCodeApplicationService {
	return &SalesVoucherCodeApplicationServiceImpl{CrudApplicationService: base}
}

type SalesVoucherCodeApplicationServiceImpl struct {
	composable.CrudApplicationService
}

func NewSalesVoucherRedemptionApplicationService(base composable.CrudApplicationService) itPromotion.SalesVoucherRedemptionApplicationService {
	return &SalesVoucherRedemptionApplicationServiceImpl{CrudApplicationService: base}
}

type SalesVoucherRedemptionApplicationServiceImpl struct {
	composable.CrudApplicationService
}
