package repository

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func NewSourcingGroupRepository(base composable.CrudRepository) itOrder.SourcingGroupCrudRepository {
	return &SourcingGroupRepositoryImpl{CrudRepository: base}
}

type SourcingGroupRepositoryImpl struct {
	composable.CrudRepository
}
