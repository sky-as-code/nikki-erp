package interfaces

import (
	// "github.com/sky-as-code/nikki-erp/modules/inventory/product/"
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func EntToProduct(entProduct *ent.Product) *Product {
	product := &Product{}
	model.MustCopy(entProduct, product)

	// Handle CommChannels relation if loaded
	// if entProduct.Edges.Variant != nil {
	// 	product.Variants = array.Map(entProduct.Edges.Variant, func(entVariant *ent.Variant) itVariant.Variant {
	// 		return *EntToVariant(entVariant)
	// 	})
	// }

	// if entProduct.Edges.Attribute != nil {

	// }

	return product
}

func EntToProducts(entProducts []*ent.Product) []Product {
	if entProducts == nil {
		return nil
	}
	return array.Map(entProducts, func(entProduct *ent.Product) Product {
		return *EntToProduct(entProduct)
	})
}

func (cmd CreateProductCommand) ToDomainModel() *Product {
	product := &Product{}
	model.MustCopy(cmd, product)
	return product
}

func (cmd UpdateProductCommand) ToDomainModel() *Product {
	product := &Product{}
	model.MustCopy(cmd, product)
	return product
}

func (this DeleteProductCommand) ToDomainModel() *Product {
	product := &Product{}
	product.Id = &this.Id
	return product
}
