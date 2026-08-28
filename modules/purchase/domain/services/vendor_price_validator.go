package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"
)

// VendorPriceValidator checks a vendor price against the three masters it references (section 23).
//
// Each reference is a plain ulid — the vendor belongs to Contacts, the product and its unit to
// Inventory and Essential — so nothing in the database stops a row naming something that does not
// exist, is archived, or belongs to a different product. A foreign key across a module boundary is
// exactly what this codebase refuses to declare, and this is the price of that: the checks are
// here, at write time, where a useful message can be produced.
//
// The ports are held rather than looked up per call, matching ProductLineValidator. Init resolves
// all three or fails, following the convention the two validators beside this one set: a module
// that booted without them would silently accept prices for parties nobody may order from, on
// products nobody may buy, in units that convert to nothing — and no such row announces itself
// until an order tries to resolve through it. The nil checks below are therefore belt-and-braces
// for a validator constructed by hand in a test, not a supported deployment.
type VendorPriceValidator struct {
	vendors  itExt.VendorExtService
	products itExt.ProductExtService
	uoms     itExt.UomExtService
}

func NewVendorPriceValidator(
	vendors itExt.VendorExtService,
	products itExt.ProductExtService,
	uoms itExt.UomExtService,
) *VendorPriceValidator {
	return &VendorPriceValidator{vendors: vendors, products: products, uoms: uoms}
}

// Validate checks one vendor price about to be written.
//
// It reports through vErrs rather than returning an error, so several problems are reported at once
// — someone correcting a bad import wants every fault in the row, not the first one and then
// another round trip.
func (this *VendorPriceValidator) Validate(
	ctx corectx.Context, record dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	assertVendorPriceSelfConsistent(record, vErrs)

	if err := this.assertVendorOrderable(ctx, record, vErrs); err != nil {
		return err
	}
	inventoryUomId, err := this.assertProductPurchasable(ctx, record, vErrs)
	if err != nil {
		return err
	}
	return this.assertPurchaseUomCompatible(ctx, record, inventoryUomId, vErrs)
}

// assertVendorOrderable refuses a price for a party that is not a usable vendor.
//
// AssertOrderable rather than a plain existence check: a party may exist, and even have a vendor
// profile, while being blocked or archived. Recording a new offer from a supplier the business has
// stopped buying from is the case this catches — and it is a quiet one, because the row would look
// perfectly ordinary until somebody tried to raise an order against it.
func (this *VendorPriceValidator) assertVendorOrderable(
	ctx corectx.Context, record dmodel.DynamicFields, vErrs *ft.ClientErrors,
) error {
	vendorId := recordString(record, models.VendorProductPriceFieldVendorId)
	if vendorId == "" || this.vendors == nil {
		return nil
	}

	result, err := this.vendors.AssertOrderable(ctx, itExt.AssertOrderableQuery{
		PartyId: model.Id(vendorId),
		OrgId:   model.Id(recordString(record, models.VendorProductPriceFieldOrgId)),
		// Field is echoed back on any violation, so the error points at this row's own column
		// rather than at whatever Contacts calls it.
		Field: models.VendorProductPriceFieldVendorId,
	})
	if err != nil {
		return errors.Wrap(err, "assertVendorOrderable")
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		// The port already phrased the refusal — whether the party is unknown, has no vendor
		// profile, or is blocked. Restating it here would put a second, vaguer sentence in front of
		// a caller who was told the specific one.
		vErrs.Append(result.ClientErrors...)
	}
	return nil
}

// assertProductPurchasable refuses a price for something that may not be bought, and returns the
// product's inventory unit for the compatibility check that follows.
//
// The variant, when given, must belong to the named template. Nothing else checks it: the two
// columns would otherwise describe different products, and resolution — which finds candidates by
// template and then prefers the variant-specific one — would return a price for a product nobody
// asked about.
func (this *VendorPriceValidator) assertProductPurchasable(
	ctx corectx.Context, record dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (string, error) {
	variantId := recordString(record, models.VendorProductPriceFieldProductVariantId)
	templateId := recordString(record, models.VendorProductPriceFieldProductTemplateId)

	// A template-wide price names no variant, so there is nothing to resolve. The template's own
	// existence is not checked here: the port reads a variant, and a template with no variants
	// cannot be bought anyway.
	if variantId == "" || this.products == nil {
		return "", nil
	}

	found, err := this.products.GetPurchasableProduct(ctx, itExt.GetPurchasableProductQuery{
		VariantId: model.Id(variantId),
	})
	if err != nil {
		return "", errors.Wrap(err, "assertProductPurchasable")
	}
	if found == nil || !found.HasData {
		appendLineViolation(vErrs, models.VendorProductPriceFieldProductVariantId,
			"purchase_vendor_product_price.product_not_found",
			"no product with id '"+variantId+"'")
		return "", nil
	}
	if !found.Data.Purchasable {
		appendLineViolation(vErrs, models.VendorProductPriceFieldProductVariantId,
			"purchase_vendor_product_price.product_not_purchasable",
			"this product is not marked as purchasable, so no vendor price applies to it")
		return "", nil
	}
	if templateId != "" && string(found.Data.TemplateId) != templateId {
		appendLineViolation(vErrs, models.VendorProductPriceFieldProductVariantId,
			"purchase_vendor_product_price.variant_template_mismatch",
			"this variant belongs to a different product than the template named on this price")
		return "", nil
	}
	return string(found.Data.InventoryUomId), nil
}

// assertPurchaseUomCompatible enforces BR-PRICE-UOM-003.
//
// The vendor's unit need not be the unit stock is counted in — a supplier sells by the carton while
// the warehouse counts bottles — but the two must share a UoM category, or the conversion that
// turns a quoted quantity into an inventory quantity is not a conversion at all. Quoting a price
// per litre for a product stocked in kilograms produces a number in the wrong dimension, and
// nothing downstream would notice.
//
// The check is a conversion attempt rather than a category comparison, deliberately: Essential owns
// what is convertible, and asking it is the only way to get the same answer it will give later
// (BR-UOM-ESS-023). Reading factors and comparing categories here would be a second implementation
// of the rule, free to disagree with the first.
func (this *VendorPriceValidator) assertPurchaseUomCompatible(
	ctx corectx.Context, record dmodel.DynamicFields, inventoryUomId string, vErrs *ft.ClientErrors,
) error {
	purchaseUomId := recordString(record, models.VendorProductPriceFieldPurchaseUomId)
	if purchaseUomId == "" || this.uoms == nil {
		return nil
	}

	found, err := this.uoms.GetUom(ctx, itExt.GetUomQuery{Id: model.Id(purchaseUomId)})
	if err != nil {
		return errors.Wrap(err, "assertPurchaseUomCompatible")
	}
	if found == nil || !found.HasData {
		appendLineViolation(vErrs, models.VendorProductPriceFieldPurchaseUomId,
			"purchase_vendor_product_price.uom_not_found",
			"no unit of measure with id '"+purchaseUomId+"'")
		return nil
	}
	// An archived unit still resolves, so historical prices keep reading; it may not appear on
	// something new.
	if found.Data.IsArchived {
		appendLineViolation(vErrs, models.VendorProductPriceFieldPurchaseUomId,
			"purchase_vendor_product_price.uom_archived",
			"this unit of measure is archived and cannot be used on a new vendor price")
		return nil
	}

	// Nothing to compare against: the price is template-wide, or the product has no stock
	// configuration — an ordinary state for a service or a non-stocked item, not an error.
	if inventoryUomId == "" || inventoryUomId == purchaseUomId {
		return nil
	}

	converted, err := this.uoms.Convert(ctx, itExt.ConvertQuantityQuery{
		Quantity:    oneUnit(),
		SourceUomId: model.Id(purchaseUomId),
		TargetUomId: model.Id(inventoryUomId),
	})
	if err != nil {
		return errors.Wrap(err, "assertPurchaseUomCompatible")
	}
	if converted != nil && converted.ClientErrors.Count() > 0 {
		appendLineViolation(vErrs, models.VendorProductPriceFieldPurchaseUomId,
			"purchase_vendor_product_price.uom_not_convertible",
			"this unit cannot be converted to the unit the product is stocked in, "+
				"so a quantity quoted in it could not be received")
	}
	return nil
}
