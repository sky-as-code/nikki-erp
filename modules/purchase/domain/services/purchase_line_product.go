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

// Product and unit-of-measure rules for a purchase line. A line keeps the quantity and unit the
// buyer typed, since that is what the vendor is being asked for; the converted value goes to
// inventory_quantity (no_update) so stock has a number in its own unit.

// ProductLineValidator applies the product and unit rules to a line.
type ProductLineValidator struct {
	products itExt.ProductExtService
	uoms     itExt.UomExtService
}

func NewProductLineValidator(
	products itExt.ProductExtService, uoms itExt.UomExtService,
) *ProductLineValidator {
	return &ProductLineValidator{products: products, uoms: uoms}
}

// PrepareLine validates a line's product and unit and computes its inventory_quantity. Product and
// unit are both optional, so absence is not an error; only a combination that makes no sense is.
// The returned string is the product's TEMPLATE id, which pricing needs immediately after this call
// to avoid a second cross-module read in the same write transaction. Empty means nothing to price
// against: a section line, a free-text charge, or a refused product.
func (this *ProductLineValidator) PrepareLine(
	ctx corectx.Context, line dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (string, error) {
	if !isMoneyBearingLine(line) {
		// A section or note buys nothing; validating one would refuse a heading for naming no
		// product.
		line[models.PurchaseOrderLineFieldInventoryQuantity] = decimal.Zero
		return "", nil
	}

	variantId := stringOf(line, models.PurchaseOrderLineFieldProductVariantId)
	quantity := decimalOf(line, models.PurchaseOrderLineFieldQuantity)
	lineUomId := stringOf(line, models.PurchaseOrderLineFieldUomId)

	if variantId == "" {
		// A priced line with no product is a legitimate free-text charge (freight, a one-off fee)
		// with nothing to convert against, so the inventory quantity is the ordered one.
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

	// The ordered quantity and unit are left exactly as the buyer typed them.
	line[models.PurchaseOrderLineFieldInventoryQuantity] = inventoryQuantity
	return templateId, nil
}

// loadPurchasableProduct resolves the variant and refuses one that cannot be bought. The three
// refusals stay distinct — bad id, not purchasable, archived — so a buyer knows which they hit.
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
	// Checked at write time, not on the read path: an archived product cannot start new business,
	// but an order that already names one must still read.
	if found.Data.Archived {
		appendLineViolation(vErrs, models.PurchaseOrderLineFieldProductVariantId,
			"purchase_order_line.product_archived",
			"this product is archived and cannot be added to a new order")
		return nil, nil
	}
	return &found.Data, nil
}

// assertUomUsable refuses an archived unit on a new line. An archived unit still resolves so
// historical lines keep reading; an unknown one is refused outright, since a typo would leave a
// quantity expressed in nothing.
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

// toInventoryQuantity converts the ordered quantity into the product's inventory unit. The
// conversion is Essential's and is never reimplemented here, or Purchase would store a number
// nobody else can reproduce. Units of different categories are refused rather than copied through:
// storing a litre quantity raw would put a volume in a mass column.
func (this *ProductLineValidator) toInventoryQuantity(
	ctx corectx.Context, quantity decimal.Decimal, lineUomId, inventoryUomId string,
	vErrs *ft.ClientErrors,
) (decimal.Decimal, error) {
	// Nothing to convert. A product with no inventory unit is an ordinary service or non-stocked
	// item, not an error.
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
		// Essential's own refusal is carried through so the caller sees which units disagreed.
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
