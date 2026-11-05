package repository

import (
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The promotion repositories build on the composable default (baserepo). The atomic usage
// counters of Confirm and Cancel (D5) are the only hand-written SQL this family will need.

func NewSalesPromotionRepository(base dynamicresource.CrudRepository) itPromotion.SalesPromotionRepository {
	return &SalesPromotionRepositoryImpl{CrudRepository: base}
}

type SalesPromotionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

// NewSalesPromotionChildRepository serves a rule, a reward, a selector, a voucher or a
// confirmed-order record: none needs more than the default yet.
func NewSalesPromotionChildRepository(
	base dynamicresource.CrudRepository,
) itPromotion.SalesPromotionChildRepository {
	return &SalesPromotionChildRepositoryImpl{CrudRepository: base}
}

type SalesPromotionChildRepositoryImpl struct {
	dynamicresource.CrudRepository
}
