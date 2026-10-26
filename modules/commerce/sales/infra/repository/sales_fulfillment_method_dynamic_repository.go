// Package repository holds the Sales repositories. Each wraps the composable default with the
// module's own type, so a layer is injected by its resource's interface rather than by a shared
// one, and a resource that later needs a bespoke query has a place to put it.
//
// All SQL in the module lives here.
package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itCatalog "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/catalog"
)

func NewSalesFulfillmentMethodRepository(base composable.CrudRepository) itCatalog.SalesFulfillmentMethodRepository {
	return &SalesFulfillmentMethodRepositoryImpl{CrudRepository: base}
}

type SalesFulfillmentMethodRepositoryImpl struct {
	composable.CrudRepository
}
