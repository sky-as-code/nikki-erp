package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ProductExtService is Purchase's read-only port onto Inventory's product catalog: whether a
// variant exists and may be bought, and what unit its stock is counted in.
type ProductExtService interface {
	// GetPurchasableProduct resolves one variant for a purchase line. HasData false means the
	// variant does not exist; a non-purchasable one comes back with Purchasable false instead, so
	// the caller can tell a bad id from a product the business does not buy.
	GetPurchasableProduct(
		ctx corectx.Context, query GetPurchasableProductQuery,
	) (*GetPurchasableProductResult, error)
}

// GetPurchasableProductQuery names the variant a line refers to.
type GetPurchasableProductQuery struct {
	VariantId model.Id
}

// GetPurchasableProductResultData is the slice of a product a purchase line depends on. It is
// deliberately not the whole variant, so Purchase cannot come to depend on fields Inventory is free
// to change.
type GetPurchasableProductResultData struct {
	VariantId  model.Id
	TemplateId model.Id

	// Purchasable is the template's purchase_ok flag, not the variant's: the decision is about the
	// product line, not one colour of it.
	Purchasable bool

	// InventoryUomId is the unit this product's stock is counted in, from Inventory's
	// stock_product_config. Empty for a service or non-stocked item, which is ordinary and not an
	// error. A purchase line's unit must share a UoM category with this one, or the converted
	// inventory_quantity would be a number in the wrong dimension.
	InventoryUomId model.Id

	// Archived reports the variant's own archived state. An archived product cannot start new
	// business but must still resolve, so orders placed before it was archived stay readable.
	Archived bool
}

type GetPurchasableProductResult struct {
	Data    GetPurchasableProductResultData
	HasData bool
}
