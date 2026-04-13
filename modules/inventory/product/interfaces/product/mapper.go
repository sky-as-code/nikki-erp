package product

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateProductCommand) ToDomainModel() *domain.Product {
	product := &domain.Product{}
	model.MustCopy(this, product)
	return product
}

func (this UpdateProductCommand) ToDomainModel() *domain.Product {
	product := &domain.Product{}
	model.MustCopy(this, product)
	return product
}
