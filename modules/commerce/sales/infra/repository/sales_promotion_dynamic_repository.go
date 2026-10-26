package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/promotion"
)

func NewSalesPromotionProgramRepository(base composable.CrudRepository) itPromotion.SalesPromotionProgramRepository {
	return &SalesPromotionProgramRepositoryImpl{CrudRepository: base}
}

type SalesPromotionProgramRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPromotionConditionGroupRepository(base composable.CrudRepository) itPromotion.SalesPromotionConditionGroupRepository {
	return &SalesPromotionConditionGroupRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionGroupRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPromotionConditionRepository(base composable.CrudRepository) itPromotion.SalesPromotionConditionRepository {
	return &SalesPromotionConditionRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPromotionConditionTargetRepository(base composable.CrudRepository) itPromotion.SalesPromotionConditionTargetRepository {
	return &SalesPromotionConditionTargetRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionTargetRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPromotionRewardRepository(base composable.CrudRepository) itPromotion.SalesPromotionRewardRepository {
	return &SalesPromotionRewardRepositoryImpl{CrudRepository: base}
}

type SalesPromotionRewardRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesPromotionCompatibilityRepository(base composable.CrudRepository) itPromotion.SalesPromotionCompatibilityRepository {
	return &SalesPromotionCompatibilityRepositoryImpl{CrudRepository: base}
}

type SalesPromotionCompatibilityRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesVoucherCodeRepository(base composable.CrudRepository) itPromotion.SalesVoucherCodeRepository {
	return &SalesVoucherCodeRepositoryImpl{CrudRepository: base}
}

type SalesVoucherCodeRepositoryImpl struct {
	composable.CrudRepository
}

func NewSalesVoucherRedemptionRepository(base composable.CrudRepository) itPromotion.SalesVoucherRedemptionRepository {
	return &SalesVoucherRedemptionRepositoryImpl{CrudRepository: base}
}

type SalesVoucherRedemptionRepositoryImpl struct {
	composable.CrudRepository
}
