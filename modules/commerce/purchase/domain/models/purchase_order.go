package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PurchaseOrderSchemaName = "purchase_order"

	PurchaseOrderFieldId                 = basemodel.FieldId
	PurchaseOrderFieldEtag               = basemodel.FieldEtag
	PurchaseOrderFieldOrgId              = basemodel.FieldOrgId
	PurchaseOrderFieldCode               = "code"
	PurchaseOrderFieldStatus             = "status"
	PurchaseOrderFieldVendorId           = "vendor_id"
	PurchaseOrderFieldVendorReference    = "vendor_reference"
	PurchaseOrderFieldSourceReference    = "source_reference"
	PurchaseOrderFieldBuyerId            = "buyer_id"
	PurchaseOrderFieldCurrencyId         = "currency_id"
	PurchaseOrderFieldOrderDeadline      = "order_deadline"
	PurchaseOrderFieldExpectedArrival    = "expected_arrival"
	PurchaseOrderFieldConfirmedAt        = "confirmed_at"
	PurchaseOrderFieldAgreementId        = "agreement_id"
	PurchaseOrderFieldSourcingGroupId    = "sourcing_group_id"
	PurchaseOrderFieldPriority           = "priority"
	PurchaseOrderFieldTermsConditions    = "terms_conditions"
	PurchaseOrderFieldIsLocked           = "is_locked"
	PurchaseOrderFieldVendorAcknowledged = "vendor_acknowledged"
	PurchaseOrderFieldUntaxedAmount      = "untaxed_amount"
	PurchaseOrderFieldTaxAmount          = "tax_amount"
	PurchaseOrderFieldTotalAmount        = "total_amount"
	PurchaseOrderFieldApprovalRequired   = "approval_required"
	PurchaseOrderFieldApprovedBy         = "approved_by"
	PurchaseOrderFieldApprovedAt         = "approved_at"
)

//go:embed purchase_order.json
var purchaseOrderSchemaJson string

func PurchaseOrderSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(purchaseOrderSchemaJson)
}

type PurchaseOrder struct {
	fields dmodel.DynamicFields
}

func NewPurchaseOrder() *PurchaseOrder {
	return &PurchaseOrder{fields: make(dmodel.DynamicFields)}
}

func NewPurchaseOrderFrom(src dmodel.DynamicFields) *PurchaseOrder {
	return &PurchaseOrder{fields: src}
}

func (this PurchaseOrder) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *PurchaseOrder) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this PurchaseOrder) GetId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldId)
}

func (this *PurchaseOrder) SetId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldId, v)
}

func (this PurchaseOrder) GetEtag() *model.Etag {
	return this.fields.GetEtag(PurchaseOrderFieldEtag)
}

func (this *PurchaseOrder) SetEtag(v *model.Etag) {
	this.fields.SetEtag(PurchaseOrderFieldEtag, v)
}

func (this PurchaseOrder) GetOrgId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldOrgId)
}

func (this *PurchaseOrder) SetOrgId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldOrgId, v)
}

func (this PurchaseOrder) GetCode() *string {
	return this.fields.GetString(PurchaseOrderFieldCode)
}

func (this *PurchaseOrder) SetCode(v *string) {
	this.fields.SetString(PurchaseOrderFieldCode, v)
}

func (this PurchaseOrder) GetStatus() *string {
	return this.fields.GetString(PurchaseOrderFieldStatus)
}

func (this *PurchaseOrder) SetStatus(v *string) {
	this.fields.SetString(PurchaseOrderFieldStatus, v)
}

func (this PurchaseOrder) GetVendorId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldVendorId)
}

func (this *PurchaseOrder) SetVendorId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldVendorId, v)
}

func (this PurchaseOrder) GetVendorReference() *string {
	return this.fields.GetString(PurchaseOrderFieldVendorReference)
}

func (this *PurchaseOrder) SetVendorReference(v *string) {
	this.fields.SetString(PurchaseOrderFieldVendorReference, v)
}

func (this PurchaseOrder) GetSourceReference() *string {
	return this.fields.GetString(PurchaseOrderFieldSourceReference)
}

func (this *PurchaseOrder) SetSourceReference(v *string) {
	this.fields.SetString(PurchaseOrderFieldSourceReference, v)
}

func (this PurchaseOrder) GetBuyerId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldBuyerId)
}

func (this *PurchaseOrder) SetBuyerId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldBuyerId, v)
}

func (this PurchaseOrder) GetCurrencyId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldCurrencyId)
}

func (this *PurchaseOrder) SetCurrencyId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldCurrencyId, v)
}

func (this PurchaseOrder) GetOrderDeadline() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PurchaseOrderFieldOrderDeadline)
}

func (this *PurchaseOrder) SetOrderDeadline(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PurchaseOrderFieldOrderDeadline, v)
}

func (this PurchaseOrder) GetExpectedArrival() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PurchaseOrderFieldExpectedArrival)
}

func (this *PurchaseOrder) SetExpectedArrival(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PurchaseOrderFieldExpectedArrival, v)
}

func (this PurchaseOrder) GetConfirmedAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PurchaseOrderFieldConfirmedAt)
}

func (this *PurchaseOrder) SetConfirmedAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PurchaseOrderFieldConfirmedAt, v)
}

func (this PurchaseOrder) GetAgreementId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldAgreementId)
}

func (this *PurchaseOrder) SetAgreementId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldAgreementId, v)
}

func (this PurchaseOrder) GetSourcingGroupId() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldSourcingGroupId)
}

func (this *PurchaseOrder) SetSourcingGroupId(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldSourcingGroupId, v)
}

func (this PurchaseOrder) GetPriority() *string {
	return this.fields.GetString(PurchaseOrderFieldPriority)
}

func (this *PurchaseOrder) SetPriority(v *string) {
	this.fields.SetString(PurchaseOrderFieldPriority, v)
}

func (this PurchaseOrder) GetTermsConditions() *string {
	return this.fields.GetString(PurchaseOrderFieldTermsConditions)
}

func (this *PurchaseOrder) SetTermsConditions(v *string) {
	this.fields.SetString(PurchaseOrderFieldTermsConditions, v)
}

func (this PurchaseOrder) GetIsLocked() *bool {
	return this.fields.GetBool(PurchaseOrderFieldIsLocked)
}

func (this *PurchaseOrder) SetIsLocked(v *bool) {
	this.fields.SetBool(PurchaseOrderFieldIsLocked, v)
}

func (this PurchaseOrder) GetVendorAcknowledged() *bool {
	return this.fields.GetBool(PurchaseOrderFieldVendorAcknowledged)
}

func (this *PurchaseOrder) SetVendorAcknowledged(v *bool) {
	this.fields.SetBool(PurchaseOrderFieldVendorAcknowledged, v)
}

func (this PurchaseOrder) GetUntaxedAmount() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderFieldUntaxedAmount)
}

func (this *PurchaseOrder) SetUntaxedAmount(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderFieldUntaxedAmount, v)
}

func (this PurchaseOrder) GetTaxAmount() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderFieldTaxAmount)
}

func (this *PurchaseOrder) SetTaxAmount(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderFieldTaxAmount, v)
}

func (this PurchaseOrder) GetTotalAmount() *decimal.Decimal {
	return this.fields.GetDecimal(PurchaseOrderFieldTotalAmount)
}

func (this *PurchaseOrder) SetTotalAmount(v *decimal.Decimal) {
	this.fields.SetDecimal(PurchaseOrderFieldTotalAmount, v)
}

func (this PurchaseOrder) GetApprovalRequired() *bool {
	return this.fields.GetBool(PurchaseOrderFieldApprovalRequired)
}

func (this *PurchaseOrder) SetApprovalRequired(v *bool) {
	this.fields.SetBool(PurchaseOrderFieldApprovalRequired, v)
}

func (this PurchaseOrder) GetApprovedBy() *model.Id {
	return this.fields.GetModelId(PurchaseOrderFieldApprovedBy)
}

func (this *PurchaseOrder) SetApprovedBy(v *model.Id) {
	this.fields.SetModelId(PurchaseOrderFieldApprovedBy, v)
}

func (this PurchaseOrder) GetApprovedAt() *model.ModelDateTime {
	return this.fields.GetModelDateTime(PurchaseOrderFieldApprovedAt)
}

func (this *PurchaseOrder) SetApprovedAt(v *model.ModelDateTime) {
	this.fields.SetModelDateTime(PurchaseOrderFieldApprovedAt, v)
}
