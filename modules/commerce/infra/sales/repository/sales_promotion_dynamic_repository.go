package repository

import (
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSalesPromotionProgramRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionProgramRepository {
	return &SalesPromotionProgramRepositoryImpl{CrudRepository: base}
}

type SalesPromotionProgramRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPromotionConditionGroupRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionConditionGroupRepository {
	return &SalesPromotionConditionGroupRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionGroupRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPromotionConditionRepository(
	base dynamicresource.CrudRepository,
) itPromotion.SalesPromotionConditionRepository {
	return &SalesPromotionConditionRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPromotionConditionTargetRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionConditionTargetRepository {
	return &SalesPromotionConditionTargetRepositoryImpl{CrudRepository: base}
}

type SalesPromotionConditionTargetRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPromotionRewardRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionRewardRepository {
	return &SalesPromotionRewardRepositoryImpl{CrudRepository: base}
}

type SalesPromotionRewardRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesPromotionCompatibilityRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionCompatibilityRepository {
	return &SalesPromotionCompatibilityRepositoryImpl{CrudRepository: base}
}

type SalesPromotionCompatibilityRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesVoucherCodeRepository(base dynamicresource.CrudRepository) itPromotion.SalesVoucherCodeRepository {
	return &SalesVoucherCodeRepositoryImpl{CrudRepository: base}
}

type SalesVoucherCodeRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func NewSalesVoucherRedemptionRepository(base dynamicresource.CrudRepository) itPromotion.SalesVoucherRedemptionRepository {
	return &SalesVoucherRedemptionRepositoryImpl{CrudRepository: base}
}

type SalesVoucherRedemptionRepositoryImpl struct {
	dynamicresource.CrudRepository
}
