package productcategory

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToProductCategory(entProductCategory *ent.ProductCategory) *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	model.MustCopy(entProductCategory, productCategory)

	// Handle relations if loaded
	// if entProductCategory.Edges.Parent != nil {
	// 	// Handle parent relation if needed
	// }

	return productCategory
}

func EntToProductCategories(entProductCategories []*ent.ProductCategory) []domain.ProductCategory {
	if entProductCategories == nil {
		return nil
	}
	return array.Map(entProductCategories, func(entProductCategory *ent.ProductCategory) domain.ProductCategory {
		return *EntToProductCategory(entProductCategory)
	})
}

func (cmd CreateProductCategoryCommand) ToDomainModel() *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	model.MustCopy(cmd, productCategory)
	return productCategory
}

func (cmd UpdateProductCategoryCommand) ToDomainModel() *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	model.MustCopy(cmd, productCategory)
	return productCategory
}

func (this DeleteProductCategoryCommand) ToDomainModel() *domain.ProductCategory {
	productCategory := &domain.ProductCategory{}
	productCategory.Id = &this.Id
	return productCategory
}
