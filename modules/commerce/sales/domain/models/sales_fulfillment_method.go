package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesFulfillmentMethodSchemaName = "sales_fulfillment_method"

	SalesFulfillmentMethodFieldId                            = "id"
	SalesFulfillmentMethodFieldOrgId                         = "org_id"
	SalesFulfillmentMethodFieldCode                          = "code"
	SalesFulfillmentMethodFieldName                          = "name"
	SalesFulfillmentMethodFieldDescription                   = "description"
	SalesFulfillmentMethodFieldFulfillmentType               = "fulfillment_type"
	SalesFulfillmentMethodFieldInitialTargetSelection        = "initial_target_selection"
	SalesFulfillmentMethodFieldMaxAttempts                   = "max_attempts"
	SalesFulfillmentMethodFieldFailureAction                 = "failure_action"
	SalesFulfillmentMethodFieldAllowTargetChange             = "allow_target_change"
	SalesFulfillmentMethodFieldAllowPartialFulfillment       = "allow_partial_fulfillment"
	SalesFulfillmentMethodFieldRequiresAuthenticatedCustomer = "requires_authenticated_customer"
	SalesFulfillmentMethodFieldReservationTtlMinutes         = "reservation_ttl_minutes"
)

// The codes of the three kiosk methods the vending flows resolve by name. They are seeded per
// organization rather than hard-coded ids, so an operator may add methods of their own; these
// constants exist because the seed and the code reading it would otherwise drift, and a lookup by a
// mistyped code fails as "no such method" rather than as a typo.
const (
	// KioskDirectAutoRefundCode is the anonymous walk-up sale: one attempt, and whatever the
	// machine failed to dispense is refunded without asking, because there is nobody to ask.
	KioskDirectAutoRefundCode = "KIOSK_DIRECT_AUTO_REFUND"

	// KioskDirectAuthenticatedCode is the same sale made by an identified customer, who keeps the
	// entitlement to what was not dispensed instead of having it refunded out from under them.
	KioskDirectAuthenticatedCode = "KIOSK_DIRECT_AUTHENTICATED"

	// KioskPickupSelectedCode is an order raised elsewhere and collected at a kiosk the customer
	// chose, which is why its target cannot be inferred from where the order was created.
	KioskPickupSelectedCode = "KIOSK_PICKUP_SELECTED"
)

//go:embed sales_fulfillment_method.json
var salesFulfillmentMethodSchemaJson string

func SalesFulfillmentMethodSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesFulfillmentMethodSchemaJson)
}

type SalesFulfillmentMethod struct {
	basemodel.DynamicModelBase
}

func NewSalesFulfillmentMethod() *SalesFulfillmentMethod {
	return &SalesFulfillmentMethod{basemodel.NewDynamicModel()}
}

func NewSalesFulfillmentMethodFrom(src dmodel.DynamicFields) *SalesFulfillmentMethod {
	return &SalesFulfillmentMethod{basemodel.NewDynamicModel(src)}
}

func (this SalesFulfillmentMethod) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesFulfillmentMethod) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesFulfillmentMethod) GetCode() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldCode)
}

func (this *SalesFulfillmentMethod) SetCode(code *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldCode, code)
}

func (this SalesFulfillmentMethod) GetName() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldName)
}

func (this *SalesFulfillmentMethod) SetName(name *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldName, name)
}

func (this SalesFulfillmentMethod) GetDescription() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldDescription)
}

func (this *SalesFulfillmentMethod) SetDescription(description *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldDescription, description)
}

func (this SalesFulfillmentMethod) GetFulfillmentType() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldFulfillmentType)
}

func (this *SalesFulfillmentMethod) SetFulfillmentType(fulfillmentType *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldFulfillmentType, fulfillmentType)
}

func (this SalesFulfillmentMethod) GetInitialTargetSelection() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldInitialTargetSelection)
}

func (this *SalesFulfillmentMethod) SetInitialTargetSelection(selection *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldInitialTargetSelection, selection)
}

func (this SalesFulfillmentMethod) GetMaxAttempts() *int32 {
	return this.GetFieldData().GetInt32(SalesFulfillmentMethodFieldMaxAttempts)
}

func (this *SalesFulfillmentMethod) SetMaxAttempts(maxAttempts *int32) {
	this.GetFieldData().SetInt32(SalesFulfillmentMethodFieldMaxAttempts, maxAttempts)
}

func (this SalesFulfillmentMethod) GetFailureAction() *string {
	return this.GetFieldData().GetString(SalesFulfillmentMethodFieldFailureAction)
}

func (this *SalesFulfillmentMethod) SetFailureAction(action *string) {
	this.GetFieldData().SetString(SalesFulfillmentMethodFieldFailureAction, action)
}

func (this SalesFulfillmentMethod) GetAllowTargetChange() *bool {
	return this.GetFieldData().GetBool(SalesFulfillmentMethodFieldAllowTargetChange)
}

func (this *SalesFulfillmentMethod) SetAllowTargetChange(allow *bool) {
	this.GetFieldData().SetBool(SalesFulfillmentMethodFieldAllowTargetChange, allow)
}

func (this SalesFulfillmentMethod) GetAllowPartialFulfillment() *bool {
	return this.GetFieldData().GetBool(SalesFulfillmentMethodFieldAllowPartialFulfillment)
}

func (this *SalesFulfillmentMethod) SetAllowPartialFulfillment(allow *bool) {
	this.GetFieldData().SetBool(SalesFulfillmentMethodFieldAllowPartialFulfillment, allow)
}

func (this SalesFulfillmentMethod) GetRequiresAuthenticatedCustomer() *bool {
	return this.GetFieldData().GetBool(SalesFulfillmentMethodFieldRequiresAuthenticatedCustomer)
}

func (this *SalesFulfillmentMethod) SetRequiresAuthenticatedCustomer(requires *bool) {
	this.GetFieldData().SetBool(SalesFulfillmentMethodFieldRequiresAuthenticatedCustomer, requires)
}

func (this SalesFulfillmentMethod) GetReservationTtlMinutes() *int32 {
	return this.GetFieldData().GetInt32(SalesFulfillmentMethodFieldReservationTtlMinutes)
}

func (this *SalesFulfillmentMethod) SetReservationTtlMinutes(minutes *int32) {
	this.GetFieldData().SetInt32(SalesFulfillmentMethodFieldReservationTtlMinutes, minutes)
}

func (this SalesFulfillmentMethod) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// IsAssignable reports whether this method may be attached to a NEW order. Archived methods stay
// readable and keep working for the fulfillments that already snapshotted them — that is the whole
// point of the snapshot — but nothing new may choose one.
func (this SalesFulfillmentMethod) IsAssignable() bool {
	archived := this.GetIsArchived()
	return archived == nil || !*archived
}

// IsExecutable reports whether this module can actually run the method's workflow. Only kiosk
// dispense is implemented; the remaining types exist in the enum so it need not be migrated when
// they are built, and every execution path asks this rather than assuming.
func (this SalesFulfillmentMethod) IsExecutable() bool {
	fulfillmentType := this.GetFulfillmentType()
	return fulfillmentType != nil &&
		FulfillmentType(*fulfillmentType) == FulfillmentTypeKioskDispense
}
