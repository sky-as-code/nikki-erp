package services

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Product and unit-of-measure rules for a purchase line (PUR-R8, BR-UOM-PUR-001..009).
//
// The central rule is that a line KEEPS what the buyer typed. The quantity and unit they entered
// are what the vendor is being asked for, and overwriting them with a converted value would change
// the order's meaning: "10 boxes" and "120 units" are the same amount of goods but not the same
// request, and only one of them is what the purchase order should say (004).
//
// The conversion lands in inventory_quantity instead, which is no_update and exists so that stock
// has a number in its own unit without the order having to lie about what was ordered (003).

// ProductLineValidator applies the product and unit rules to a line.
//
// It holds its two ports rather than resolving them per call: they are resolved once at Init, and a
// validator that could not find them would be a wiring fault rather than a request problem.
type ProductLineValidator struct {
	products itExt.ProductExtService
	uoms     itExt.UomExtService
}

func NewProductLineValidator(
	products itExt.ProductExtService, uoms itExt.UomExtService,
) *ProductLineValidator {
	return &ProductLineValidator{products: products, uoms: uoms}
}

// PrepareLine validates a line's product and unit and computes its inventory_quantity.
//
// It returns the value to store, or a refusal. Both the product and the unit are optional on a
// line — a note or a section has neither, and a product line for a one-off service may name a
// product with no stock configuration — so absence is not itself an error. What is an error is a
// combination that cannot be made sense of.
// It also returns the product's TEMPLATE id when it resolved one. Pricing needs it, and pricing
// runs immediately after this call — re-reading Inventory for an id already in hand would be a
// second cross-module round trip inside the same write transaction. Empty means there is nothing
// to price against: a section line, a free-text charge, or a product that was refused.
func (this *ProductLineValidator) PrepareLine(
	ctx corectx.Context, line dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (string, error) {
	if !isMoneyBearingLine(line) {
		// A section or a note buys nothing, so it has no product to check and no quantity to
		// convert. Validating one would refuse a heading for not naming a product.
		line[models.PurchaseOrderLineFieldInventoryQuantity] = decimal.Zero
		return "", nil
	}

	variantId := stringOf(line, models.PurchaseOrderLineFieldProductVariantId)
	quantity := decimalOf(line, models.PurchaseOrderLineFieldQuantity)
	lineUomId := stringOf(line, models.PurchaseOrderLineFieldUomId)

	if variantId == "" {
		// A priced line with no product is a free-text charge — freight, a one-off fee. It is
		// legitimate, and there is nothing to convert against, so the inventory quantity is the
		// ordered quantity unchanged.
		line[models.PurchaseOrderLineFieldInventoryQuantity] = quantity
		return "", nil
	}

	product, err := this.loadPurchasableProduct(ctx, variantId, vErrs)
	if err != nil || product == nil {
		return "", err
	}
	templateId := string(product.TemplateId)

	if err := this.assertUomUsable(ctx, lineUomId, vErrs); err != nil {
		return templateId, err
	}
	if vErrs.Count() > 0 {
		return templateId, nil
	}

	inventoryQuantity, err := this.toInventoryQuantity(
		ctx, quantity, lineUomId, string(product.InventoryUomId), vErrs)
	if err != nil {
		return templateId, err
	}
	if vErrs.Count() > 0 {
		return templateId, nil
	}

	// The ordered quantity and unit are left exactly as the buyer typed them (004).
	line[models.PurchaseOrderLineFieldInventoryQuantity] = inventoryQuantity
	return templateId, nil
}

// loadPurchasableProduct resolves the variant and refuses one that cannot be bought.
//
// The three refusals are deliberately distinct. "No such product" is a bad id; "not purchasable" is
// a real product the business has decided not to buy (D4); "archived" is a product that was bought
// before and is not bought now. Collapsing them into one message would leave a buyer guessing which
// of the three they hit.
func (this *ProductLineValidator) loadPurchasableProduct(
	ctx corectx.Context, variantId string, vErrs *ft.ClientErrors,
) (*itExt.GetPurchasableProductResultData, error) {
	found, err := this.products.GetPurchasableProduct(ctx, itExt.GetPurchasableProductQuery{
		VariantId: model.Id(variantId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadPurchasableProduct")
	}
	if found == nil || !found.HasData {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldProductVariantId,
			"purchase_order_line.product_not_found",
			"no product with id '"+variantId+"'")
		return nil, nil
	}

	if !found.Data.Purchasable {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldProductVariantId,
			"purchase_order_line.product_not_purchasable",
			"this product is not marked as purchasable and cannot be ordered")
		return nil, nil
	}
	// An archived product cannot start new business, but an order that already names one must
	// still read (008) — which is why this check is here, at write time, and not on the read path.
	if found.Data.Archived {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldProductVariantId,
			"purchase_order_line.product_archived",
			"this product is archived and cannot be added to a new order")
		return nil, nil
	}
	return &found.Data, nil
}

// assertUomUsable refuses an archived unit on a new line (008).
//
// A unit that has been archived is still resolvable, so historical lines keep reading; what it may
// not do is appear on something new. An unknown unit is refused outright: a typo would otherwise
// produce a line whose quantity is expressed in nothing.
func (this *ProductLineValidator) assertUomUsable(
	ctx corectx.Context, uomId string, vErrs *ft.ClientErrors,
) error {
	if uomId == "" {
		return nil
	}

	found, err := this.uoms.GetUom(ctx, itExt.GetUomQuery{Id: model.Id(uomId)})
	if err != nil {
		return errors.Wrap(err, "assertUomUsable")
	}
	if found == nil || !found.HasData {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldUomId,
			"purchase_order_line.uom_not_found",
			"no unit of measure with id '"+uomId+"'")
		return nil
	}
	if found.Data.IsArchived {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldUomId,
			"purchase_order_line.uom_archived",
			"this unit of measure is archived and cannot be used on a new line")
	}
	return nil
}

// toInventoryQuantity converts the ordered quantity into the product's inventory unit (003, 009).
//
// The conversion itself is Essential's, never reimplemented here: reading factors and dividing them
// would give a second answer to the same question, and the one Purchase stored would be the one
// nobody could reproduce (BR-UOM-ESS-023).
//
// The cross-category refusal is the point of 009. Ordering in litres a product whose stock is
// counted in kilograms is not a conversion Essential can do, and storing the raw number would put a
// mass in a volume column.
func (this *ProductLineValidator) toInventoryQuantity(
	ctx corectx.Context, quantity decimal.Decimal, lineUomId, inventoryUomId string,
	vErrs *ft.ClientErrors,
) (decimal.Decimal, error) {
	// No inventory unit configured, or the line is already in it: nothing to convert. A product
	// with no stock configuration is an ordinary service or non-stocked item.
	if inventoryUomId == "" || lineUomId == "" || lineUomId == inventoryUomId {
		return quantity, nil
	}

	converted, err := this.uoms.Convert(ctx, itExt.ConvertQuantityQuery{
		Quantity:    quantity,
		SourceUomId: model.Id(lineUomId),
		TargetUomId: model.Id(inventoryUomId),
	})
	if err != nil {
		return decimal.Zero, errors.Wrap(err, "toInventoryQuantity")
	}
	if converted != nil && converted.ClientErrors.Count() > 0 {
		// Essential refused, which for two units of different categories is exactly what it should
		// do. Its reasons are carried through rather than restated, so the caller sees which units
		// disagreed instead of a generic refusal.
		vErrs.ConcatPtr(&converted.ClientErrors)
		return decimal.Zero, nil
	}
	if converted == nil || !converted.HasData {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldUomId,
			"purchase_order_line.uom_not_convertible",
			"this line's unit cannot be converted to the product's inventory unit; "+
				"they must belong to the same unit-of-measure category")
		return decimal.Zero, nil
	}
	return converted.Data.Quantity, nil
}

func appendLineViolation(vErrs *ft.ClientErrors, field, key, message string) {
	vErrs.Append(*ft.NewBusinessViolation(field, key, message))
}
