package repository

import (
	itOrder "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/order"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func NewSourcingGroupRepository(base dynamicresource.CrudRepository) itOrder.SourcingGroupCrudRepository {
	return &SourcingGroupRepositoryImpl{CrudRepository: base}
}

type SourcingGroupRepositoryImpl struct {
	dynamicresource.CrudRepository
}
