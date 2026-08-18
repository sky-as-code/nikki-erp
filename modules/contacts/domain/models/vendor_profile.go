package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type VendorStatus string

const (
	VendorStatusProposed    = VendorStatus("proposed")
	VendorStatusActive      = VendorStatus("active")
	VendorStatusSuspended   = VendorStatus("suspended")
	VendorStatusBlacklisted = VendorStatus("blacklisted")
)

// VendorProfile is the supplier-specific facts about a party.
//
// It is a separate table rather than columns on contacts_party, and that is the whole point of the
// design. The dynamic model system has no discriminator and no single-table inheritance, so a
// type = 'vendor' enum value would give zero validation leverage: payment_terms and lead_time_days
// would be nullable-and-unvalidated on every individual and company row, with nothing able to
// enforce "required when vendor".
//
// As its own table the question "is this party a vendor?" becomes "does it have a profile row?",
// which is checkable. It also lets one party be a supplier and a customer at once — a customer
// profile would be a second sidecar — without the duplicate contact records that separate Vendor
// and Customer tables force.
//
// The shape follows inventory_stock_product_config, which is the same idea: one module's settings
// hanging off another module's master record, keyed one-to-one by a composite unique on
// (owner_id, org_id) rather than by a declared one:one edge.
//
// Purchase reads this through a port and never imports this package.
const (
	VendorProfileSchemaName = "contacts_vendor_profile"

	VendorProfileFieldId                = basemodel.FieldId
	VendorProfileFieldEtag              = basemodel.FieldEtag
	VendorProfileFieldOrgId             = basemodel.FieldOrgId
	VendorProfileFieldIsArchived        = basemodel.FieldIsArchived
	VendorProfileFieldPartyId           = "party_id"
	VendorProfileFieldStatus            = "status"
	VendorProfileFieldStatusReason      = "status_reason"
	VendorProfileFieldDefaultCurrencyId = "default_currency_id"
	VendorProfileFieldPaymentTerms      = "payment_terms"
	VendorProfileFieldLeadTimeDays      = "lead_time_days"
	VendorProfileFieldNote              = "note"

	VendorProfileEdgeParty = "party"
)

//go:embed vendor_profile.json
var vendorProfileSchemaJson string

func VendorProfileSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(vendorProfileSchemaJson)
}

type VendorProfile struct {
	fields dmodel.DynamicFields
}

func NewVendorProfile() *VendorProfile {
	return &VendorProfile{fields: make(dmodel.DynamicFields)}
}

func NewVendorProfileFrom(src dmodel.DynamicFields) *VendorProfile {
	return &VendorProfile{fields: src}
}

func (this VendorProfile) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *VendorProfile) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this VendorProfile) GetId() *model.Id {
	return this.fields.GetModelId(VendorProfileFieldId)
}

func (this *VendorProfile) SetId(v *model.Id) {
	this.fields.SetModelId(VendorProfileFieldId, v)
}

func (this VendorProfile) GetEtag() *model.Etag {
	return this.fields.GetEtag(VendorProfileFieldEtag)
}

func (this *VendorProfile) SetEtag(v *model.Etag) {
	this.fields.SetEtag(VendorProfileFieldEtag, v)
}

func (this VendorProfile) GetOrgId() *model.Id {
	return this.fields.GetModelId(VendorProfileFieldOrgId)
}

func (this *VendorProfile) SetOrgId(v *model.Id) {
	this.fields.SetModelId(VendorProfileFieldOrgId, v)
}

func (this VendorProfile) IsArchived() *bool {
	return this.fields.GetBool(VendorProfileFieldIsArchived)
}

func (this *VendorProfile) SetIsArchived(v *bool) {
	this.fields.SetBool(VendorProfileFieldIsArchived, v)
}

func (this VendorProfile) GetPartyId() *model.Id {
	return this.fields.GetModelId(VendorProfileFieldPartyId)
}

func (this *VendorProfile) SetPartyId(v *model.Id) {
	this.fields.SetModelId(VendorProfileFieldPartyId, v)
}

func (this VendorProfile) GetStatus() *string {
	return this.fields.GetString(VendorProfileFieldStatus)
}

func (this *VendorProfile) SetStatus(v *string) {
	this.fields.SetString(VendorProfileFieldStatus, v)
}

func (this VendorProfile) GetStatusReason() *string {
	return this.fields.GetString(VendorProfileFieldStatusReason)
}

func (this *VendorProfile) SetStatusReason(v *string) {
	this.fields.SetString(VendorProfileFieldStatusReason, v)
}

func (this VendorProfile) GetDefaultCurrencyId() *model.Id {
	return this.fields.GetModelId(VendorProfileFieldDefaultCurrencyId)
}

func (this *VendorProfile) SetDefaultCurrencyId(v *model.Id) {
	this.fields.SetModelId(VendorProfileFieldDefaultCurrencyId, v)
}

func (this VendorProfile) GetPaymentTerms() *string {
	return this.fields.GetString(VendorProfileFieldPaymentTerms)
}

func (this *VendorProfile) SetPaymentTerms(v *string) {
	this.fields.SetString(VendorProfileFieldPaymentTerms, v)
}

func (this VendorProfile) GetLeadTimeDays() *int32 {
	return this.fields.GetInt32(VendorProfileFieldLeadTimeDays)
}

func (this *VendorProfile) SetLeadTimeDays(v *int32) {
	this.fields.SetInt32(VendorProfileFieldLeadTimeDays, v)
}

func (this VendorProfile) GetNote() *string {
	return this.fields.GetString(VendorProfileFieldNote)
}

func (this *VendorProfile) SetNote(v *string) {
	this.fields.SetString(VendorProfileFieldNote, v)
}

// IsOrderable reports whether a new purchase order may name this vendor.
//
// Only an active, unarchived profile qualifies. Proposed means qualification is unfinished,
// suspended and blacklisted are deliberate blocks, and an archived profile is out of the working
// set — but all four stay readable, because orders already placed against them must keep resolving
// their vendor.
func (this VendorProfile) IsOrderable() bool {
	if archived := this.IsArchived(); archived != nil && *archived {
		return false
	}
	status := this.GetStatus()
	return status != nil && *status == string(VendorStatusActive)
}
