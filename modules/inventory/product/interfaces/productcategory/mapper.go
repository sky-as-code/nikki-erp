package productcategory

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func (this CreateProductCategoryCommand) ToDomainModel() *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	model.MustCopy(this, productCategory)
	return productCategory
}

func (this UpdateProductCategoryCommand) ToDomainModel() *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	model.MustCopy(this, productCategory)
	return productCategory
}
