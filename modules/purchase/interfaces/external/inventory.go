package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// ProductExtService is Purchase's port onto Inventory's product catalog.
//
// It answers the two questions a purchase line needs and nothing else: does this variant exist and
// may it be bought, and what unit is its stock counted in. Purchase must not be able to manage
// master data through this port — writes stay with Inventory.
type ProductExtService interface {
	// GetPurchasableProduct resolves one variant for a purchase line.
	//
	// HasData false means the variant does not exist. A variant that exists but whose template is
	// not purchasable comes back with Purchasable false rather than as an error, so the caller can
	// report the two cases differently: one is a bad id, the other is a real product the business
	// has decided not to buy.
	GetPurchasableProduct(
		ctx corectx.Context, query GetPurchasableProductQuery,
	) (*GetPurchasableProductResult, error)
}

// GetPurchasableProductQuery names the variant a line refers to.
type GetPurchasableProductQuery struct {
	VariantId model.Id
}

// GetPurchasableProductResultData is the slice of a product a purchase line depends on.
//
// It is deliberately not the whole variant. A consumer holding the full record would start reading
// fields that Inventory is free to change, and the port would stop being a contract.
type GetPurchasableProductResultData struct {
	VariantId  model.Id
	TemplateId model.Id

	// Purchasable is the template's purchase_ok flag (D4). It lives on the template rather than
	// the variant because "we buy this product" is a decision about the product line, not about
	// one colour of it.
	Purchasable bool

	// InventoryUomId is the unit this product's stock is counted in, from Inventory's own
	// stock_product_config. It is empty when the product has no stock configuration — a perfectly
	// ordinary state for a service or a non-stocked item, and NOT an error.
	//
	// PUR-R8 checks a purchase line's unit against this one: the two must share a UoM category, or
	// the converted inventory_quantity would be a number in the wrong dimension.
	InventoryUomId model.Id

	// Archived reports the variant's own archived state. An archived product cannot start new
	// business but must still resolve, so that an order placed before it was archived stays
	// readable (the same rule Inventory applies to its operation types).
	Archived bool
}

// GetPurchasableProductResult carries the data alongside HasData, in the shape every other port in
// this codebase uses.
type GetPurchasableProductResult struct {
	Data    GetPurchasableProductResultData
	HasData bool
}
