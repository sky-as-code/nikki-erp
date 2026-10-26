package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itOrder "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/order"
)

func NewSourcingGroupRepository(base composable.CrudRepository) itOrder.SourcingGroupCrudRepository {
	return &SourcingGroupRepositoryImpl{CrudRepository: base}
}

type SourcingGroupRepositoryImpl struct {
	composable.CrudRepository
}
