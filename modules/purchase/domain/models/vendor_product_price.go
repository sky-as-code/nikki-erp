package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// What a vendor currently offers a product at.
//
// The resource this change request adds, and the answer to a question the system previously could
// not express: "what does this supplier charge for this product, in what unit, from what quantity,
// and until when". Before it, the nearest thing was a purchase agreement line — but an agreement is
// a commitment negotiated for a period, while this is a standing offer, and conflating the two
// would mean a price could not be recorded without pretending a contract existed.
//
// It is emphatically NOT the product's cost (PRICE-INV-004, PRICE-INV-009). A vendor's offer and
// what the business has actually paid and valued are different numbers and routinely differ: a
// product costing 10,200 may be offered at 10,000 by one vendor and 9,500 by another. Cost is
// Inventory's to calculate from goods actually received; this is what somebody is asking.
const (
	VendorProductPriceSchemaName = "purchase_vendor_product_price"

	VendorProductPriceFieldId                = basemodel.FieldId
	VendorProductPriceFieldEtag              = basemodel.FieldEtag
	VendorProductPriceFieldOrgId             = basemodel.FieldOrgId
	VendorProductPriceFieldIsArchived        = basemodel.FieldIsArchived
	VendorProductPriceFieldVendorId          = "vendor_id"
	VendorProductPriceFieldProductTemplateId = "product_template_id"
	VendorProductPriceFieldProductVariantId  = "product_variant_id"
	VendorProductPriceFieldPurchaseUomId     = "purchase_uom_id"
	VendorProductPriceFieldCurrencyId        = "currency_id"
	VendorProductPriceFieldMinQuantity       = "min_quantity"
	VendorProductPriceFieldUnitPrice         = "unit_price"
	VendorProductPriceFieldValidFrom         = "valid_from"
	VendorProductPriceFieldValidTo           = "valid_to"
	VendorProductPriceFieldLeadTimeDays      = "lead_time_days"
	VendorProductPriceFieldSequence          = "sequence"
	VendorProductPriceFieldVendorProductCode = "vendor_product_code"
	VendorProductPriceFieldVendorProductName = "vendor_product_name"
)

//go:embed vendor_product_price.json
var vendorProductPriceSchemaJson string

func VendorProductPriceSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(vendorProductPriceSchemaJson)
}

type VendorProductPrice struct {
	fields dmodel.DynamicFields
}

func NewVendorProductPrice() *VendorProductPrice {
	return &VendorProductPrice{fields: make(dmodel.DynamicFields)}
}

func NewVendorProductPriceFrom(src dmodel.DynamicFields) *VendorProductPrice {
	return &VendorProductPrice{fields: src}
}

func (this VendorProductPrice) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *VendorProductPrice) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this VendorProductPrice) GetId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldId)
}

func (this *VendorProductPrice) SetId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldId, v)
}

func (this VendorProductPrice) GetEtag() *model.Etag {
	return this.fields.GetEtag(VendorProductPriceFieldEtag)
}

func (this *VendorProductPrice) SetEtag(v *model.Etag) {
	this.fields.SetEtag(VendorProductPriceFieldEtag, v)
}

func (this VendorProductPrice) GetOrgId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldOrgId)
}

func (this *VendorProductPrice) SetOrgId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldOrgId, v)
}

func (this VendorProductPrice) GetVendorId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldVendorId)
}

func (this *VendorProductPrice) SetVendorId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldVendorId, v)
}

func (this VendorProductPrice) GetProductTemplateId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldProductTemplateId)
}

func (this *VendorProductPrice) SetProductTemplateId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldProductTemplateId, v)
}

func (this VendorProductPrice) GetProductVariantId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldProductVariantId)
}

func (this *VendorProductPrice) SetProductVariantId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldProductVariantId, v)
}

func (this VendorProductPrice) GetPurchaseUomId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldPurchaseUomId)
}

func (this *VendorProductPrice) SetPurchaseUomId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldPurchaseUomId, v)
}

func (this VendorProductPrice) GetCurrencyId() *model.Id {
	return this.fields.GetModelId(VendorProductPriceFieldCurrencyId)
}

func (this *VendorProductPrice) SetCurrencyId(v *model.Id) {
	this.fields.SetModelId(VendorProductPriceFieldCurrencyId, v)
}

func (this VendorProductPrice) GetMinQuantity() *decimal.Decimal {
	return this.fields.GetDecimal(VendorProductPriceFieldMinQuantity)
}

func (this *VendorProductPrice) SetMinQuantity(v *decimal.Decimal) {
	this.fields.SetDecimal(VendorProductPriceFieldMinQuantity, v)
}

func (this VendorProductPrice) GetUnitPrice() *decimal.Decimal {
	return this.fields.GetDecimal(VendorProductPriceFieldUnitPrice)
}

func (this *VendorProductPrice) SetUnitPrice(v *decimal.Decimal) {
	this.fields.SetDecimal(VendorProductPriceFieldUnitPrice, v)
}

func (this VendorProductPrice) GetValidFrom() *model.ModelDateTime {
	return this.fields.GetModelDateTime(VendorProductPriceFieldValidFrom)
}

func (this *VendorProductPrice) SetValidFrom(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(VendorProductPriceFieldValidFrom, v)
}

func (this VendorProductPrice) GetValidTo() *model.ModelDateTime {
	return this.fields.GetModelDateTime(VendorProductPriceFieldValidTo)
}

func (this *VendorProductPrice) SetValidTo(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(VendorProductPriceFieldValidTo, v)
}

func (this VendorProductPrice) GetLeadTimeDays() *int32 {
	return this.fields.GetInt32(VendorProductPriceFieldLeadTimeDays)
}

func (this *VendorProductPrice) SetLeadTimeDays(v *int32) {
	this.fields.SetInt32(VendorProductPriceFieldLeadTimeDays, v)
}

func (this VendorProductPrice) GetSequence() *int32 {
	return this.fields.GetInt32(VendorProductPriceFieldSequence)
}

func (this *VendorProductPrice) SetSequence(v *int32) {
	this.fields.SetInt32(VendorProductPriceFieldSequence, v)
}

func (this VendorProductPrice) GetVendorProductCode() *string {
	return this.fields.GetString(VendorProductPriceFieldVendorProductCode)
}

func (this *VendorProductPrice) SetVendorProductCode(v *string) {
	this.fields.SetString(VendorProductPriceFieldVendorProductCode, v)
}

func (this VendorProductPrice) GetVendorProductName() *string {
	return this.fields.GetString(VendorProductPriceFieldVendorProductName)
}

func (this *VendorProductPrice) SetVendorProductName(v *string) {
	this.fields.SetString(VendorProductPriceFieldVendorProductName, v)
}
