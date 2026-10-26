package models

import (
	_ "embed"
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesOrderFulfillmentSchemaName = "sales_order_fulfillment"

	SalesOrderFulfillmentFieldId                      = "id"
	SalesOrderFulfillmentFieldOrgId                   = "org_id"
	SalesOrderFulfillmentFieldSalesOrderId            = "sales_order_id"
	SalesOrderFulfillmentFieldFulfillmentMethodId     = "fulfillment_method_id"
	SalesOrderFulfillmentFieldFulfillmentType         = "fulfillment_type"
	SalesOrderFulfillmentFieldTargetOutletId          = "target_outlet_id"
	SalesOrderFulfillmentFieldFulfillmentStatus       = "fulfillment_status"
	SalesOrderFulfillmentFieldMaxAttempts             = "max_attempts"
	SalesOrderFulfillmentFieldFailureAction           = "failure_action"
	SalesOrderFulfillmentFieldAllowTargetChange       = "allow_target_change"
	SalesOrderFulfillmentFieldAllowPartialFulfillment = "allow_partial_fulfillment"
	SalesOrderFulfillmentFieldReservationTtlMinutes   = "reservation_ttl_minutes"
	SalesOrderFulfillmentFieldReservationExpiresAt    = "reservation_expires_at"
)

//go:embed sales_order_fulfillment.json
var salesOrderFulfillmentSchemaJson string

func SalesOrderFulfillmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesOrderFulfillmentSchemaJson)
}

type SalesOrderFulfillment struct {
	basemodel.DynamicModelBase
}

func NewSalesOrderFulfillment() *SalesOrderFulfillment {
	return &SalesOrderFulfillment{basemodel.NewDynamicModel()}
}

func NewSalesOrderFulfillmentFrom(src dmodel.DynamicFields) *SalesOrderFulfillment {
	return &SalesOrderFulfillment{basemodel.NewDynamicModel(src)}
}

func (this SalesOrderFulfillment) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this *SalesOrderFulfillment) SetId(id *model.Id) {
	this.GetFieldData().SetModelId(basemodel.FieldId, id)
}

func (this SalesOrderFulfillment) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentFieldOrgId)
}

func (this *SalesOrderFulfillment) SetOrgId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentFieldOrgId, value)
}

func (this SalesOrderFulfillment) GetSalesOrderId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentFieldSalesOrderId)
}

func (this *SalesOrderFulfillment) SetSalesOrderId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentFieldSalesOrderId, value)
}

func (this SalesOrderFulfillment) GetFulfillmentMethodId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentFieldFulfillmentMethodId)
}

func (this *SalesOrderFulfillment) SetFulfillmentMethodId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentFieldFulfillmentMethodId, value)
}

func (this SalesOrderFulfillment) GetFulfillmentType() *string {
	return this.GetFieldData().GetString(SalesOrderFulfillmentFieldFulfillmentType)
}

func (this *SalesOrderFulfillment) SetFulfillmentType(value *string) {
	this.GetFieldData().SetString(SalesOrderFulfillmentFieldFulfillmentType, value)
}

func (this SalesOrderFulfillment) GetTargetOutletId() *model.Id {
	return this.GetFieldData().GetModelId(SalesOrderFulfillmentFieldTargetOutletId)
}

func (this *SalesOrderFulfillment) SetTargetOutletId(value *model.Id) {
	this.GetFieldData().SetModelId(SalesOrderFulfillmentFieldTargetOutletId, value)
}

func (this SalesOrderFulfillment) GetFulfillmentStatus() *string {
	return this.GetFieldData().GetString(SalesOrderFulfillmentFieldFulfillmentStatus)
}

func (this *SalesOrderFulfillment) SetFulfillmentStatus(value *string) {
	this.GetFieldData().SetString(SalesOrderFulfillmentFieldFulfillmentStatus, value)
}

func (this SalesOrderFulfillment) GetMaxAttempts() *int32 {
	return this.GetFieldData().GetInt32(SalesOrderFulfillmentFieldMaxAttempts)
}

func (this *SalesOrderFulfillment) SetMaxAttempts(value *int32) {
	this.GetFieldData().SetInt32(SalesOrderFulfillmentFieldMaxAttempts, value)
}

func (this SalesOrderFulfillment) GetFailureAction() *string {
	return this.GetFieldData().GetString(SalesOrderFulfillmentFieldFailureAction)
}

func (this *SalesOrderFulfillment) SetFailureAction(value *string) {
	this.GetFieldData().SetString(SalesOrderFulfillmentFieldFailureAction, value)
}

func (this SalesOrderFulfillment) GetAllowTargetChange() *bool {
	return this.GetFieldData().GetBool(SalesOrderFulfillmentFieldAllowTargetChange)
}

func (this *SalesOrderFulfillment) SetAllowTargetChange(value *bool) {
	this.GetFieldData().SetBool(SalesOrderFulfillmentFieldAllowTargetChange, value)
}

func (this SalesOrderFulfillment) GetAllowPartialFulfillment() *bool {
	return this.GetFieldData().GetBool(SalesOrderFulfillmentFieldAllowPartialFulfillment)
}

func (this *SalesOrderFulfillment) SetAllowPartialFulfillment(value *bool) {
	this.GetFieldData().SetBool(SalesOrderFulfillmentFieldAllowPartialFulfillment, value)
}

func (this SalesOrderFulfillment) GetReservationTtlMinutes() *int32 {
	return this.GetFieldData().GetInt32(SalesOrderFulfillmentFieldReservationTtlMinutes)
}

func (this *SalesOrderFulfillment) SetReservationTtlMinutes(value *int32) {
	this.GetFieldData().SetInt32(SalesOrderFulfillmentFieldReservationTtlMinutes, value)
}

func (this SalesOrderFulfillment) GetReservationExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesOrderFulfillmentFieldReservationExpiresAt)
}

func (this *SalesOrderFulfillment) SetReservationExpiresAt(value *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(SalesOrderFulfillmentFieldReservationExpiresAt, value)
}

// IsReservable reports that a reservation may be placed or replaced for this fulfillment. Expired
// is included deliberately: an expired fulfillment lost its stock but not its entitlement, so the
// customer may reserve again at the same or another target.
func (this SalesOrderFulfillment) IsReservable() bool {
	switch FulfillmentStatus(derefString(this.GetFulfillmentStatus())) {
	case FulfillmentStatusPendingReservation,
		FulfillmentStatusReserved,
		FulfillmentStatusReady,
		FulfillmentStatusExpired,
		FulfillmentStatusWaitingCustomerAction,
		FulfillmentStatusPartiallyFulfilled:
		return true
	}
	return false
}

// IsTerminal reports that nothing further will happen to this fulfillment on its own. Note that
// waiting_customer_action is NOT terminal: the goods are still owed and the customer may still act.
func (this SalesOrderFulfillment) IsTerminal() bool {
	switch FulfillmentStatus(derefString(this.GetFulfillmentStatus())) {
	case FulfillmentStatusCompleted, FulfillmentStatusCancelled:
		return true
	}
	return false
}

// IsAttemptable reports that an executor may be asked to hand goods over now. in_progress is absent
// on purpose: a second attempt against an outstanding one would dispense the same reservation
// twice, handing over goods that were paid for once.
func (this SalesOrderFulfillment) IsAttemptable() bool {
	switch FulfillmentStatus(derefString(this.GetFulfillmentStatus())) {
	case FulfillmentStatusReady,
		FulfillmentStatusReserved,
		FulfillmentStatusPartiallyFulfilled,
		FulfillmentStatusWaitingCustomerAction:
		return true
	}
	return false
}

// ExecutesKioskDispense reports whether this fulfillment runs the one workflow that exists. Every
// execution path asks rather than assumes, so that a method carrying one of the four reserved types
// is refused outright instead of half-running against rules written for a vending machine.
func (this SalesOrderFulfillment) ExecutesKioskDispense() bool {
	return FulfillmentType(derefString(this.GetFulfillmentType())) == FulfillmentTypeKioskDispense
}

// HasLapsed reports that the reservation deadline has passed. A fulfillment with no TTL never
// lapses, which is what a sale dispensed seconds after payment wants.
func (this SalesOrderFulfillment) HasLapsed(asOf time.Time) bool {
	expiresAt := this.GetReservationExpiresAt()
	if expiresAt == nil {
		return false
	}
	return expiresAt.GoTime().Before(asOf)
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
