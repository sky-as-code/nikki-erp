package product

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

func EntToProduct(entProduct *ent.Product) *domain.Product {
	product := &domain.Product{}
	model.MustCopy(entProduct, product)

	// Handle CommChannels relation if loaded
	// if entProduct.Edges.Variant != nil {
	// 	product.Variants = array.Map(entProduct.Edges.Variant, func(entVariant *ent.Variant) domain.Variant {
	// 		return *EntToVariant(entVariant)
	// 	})
	// }

	// if entProduct.Edges.Attribute != nil {

	// }

	return product
}

func EntToProducts(entProducts []*ent.Product) []domain.Product {
	if entProducts == nil {
		return nil
	}
	return array.Map(entProducts, func(entProduct *ent.Product) domain.Product {
		return *EntToProduct(entProduct)
	})
}

func (cmd CreateProductCommand) ToDomainModel() *domain.Product {
	product := &domain.Product{}
	model.MustCopy(cmd, product)
	return product
}

func (cmd UpdateProductCommand) ToDomainModel() *domain.Product {
	product := &domain.Product{}
	model.MustCopy(cmd, product)
	return product
}

func (this DeleteProductCommand) ToDomainModel() *domain.Product {
	product := &domain.Product{}
	product.Id = &this.Id
	return product
}
