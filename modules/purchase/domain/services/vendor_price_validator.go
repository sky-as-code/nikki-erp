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

// VendorPriceValidator checks a vendor price against the vendor, product and unit masters it
// references. Those references are plain ulids owned by other modules, so no foreign key stops a
// row naming something missing, archived, or belonging to a different product; the checks happen
// here at write time. The nil port checks below cover a validator built by hand in a test, not a
// supported deployment.
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

// Validate checks one vendor price about to be written. Problems go to vErrs rather than a returned
// error so every fault in the row is reported in one pass.
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

// assertVendorOrderable refuses a price for a party that is not a usable vendor. AssertOrderable
// rather than an existence check, because a party may exist and have a vendor profile while being
// blocked or archived.
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
		// Echoed back on any violation so the error names this row's column, not the Contacts one.
		Field: models.VendorProductPriceFieldVendorId,
	})
	if err != nil {
		return errors.Wrap(err, "assertVendorOrderable")
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		// The port already phrased the specific refusal; restating it would only make it vaguer.
		vErrs.Append(result.ClientErrors...)
	}
	return nil
}

// assertProductPurchasable refuses a price for something that may not be bought, and returns the
// product's inventory unit for the compatibility check that follows. A given variant must belong to
// the named template; nothing else checks that, and resolution (which finds candidates by template
// then prefers the variant-specific one) would otherwise return a price for the wrong product.
func (this *VendorPriceValidator) assertProductPurchasable(
	ctx corectx.Context, record dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (string, error) {
	variantId := recordString(record, models.VendorProductPriceFieldProductVariantId)
	templateId := recordString(record, models.VendorProductPriceFieldProductTemplateId)

	// A template-wide price names no variant, so there is nothing to resolve. The template's own
	// existence is not checked: the port reads a variant, and a template with no variants cannot be
	// bought anyway.
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

// assertPurchaseUomCompatible requires the vendor's unit to be convertible to the unit stock is
// counted in. The two need not match (a supplier sells cartons, the warehouse counts bottles), but
// they must share a UoM category, or a quoted quantity converts into the wrong dimension and
// nothing downstream notices. The check is an actual conversion attempt rather than a category
// comparison so that Essential, which owns convertibility, gives the answer instead of a second
// implementation of the rule here.
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
