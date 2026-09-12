package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// DeliveryStatus is the transport state of one attempt on one channel. It is never read state:
// a delivery may be 'sent' while the notification is still unread, and a 'failed' one may sit
// against a notification the user has already read elsewhere (BR 26).
type DeliveryStatus string

const (
	DeliveryStatusPending    = DeliveryStatus("pending")
	DeliveryStatusProcessing = DeliveryStatus("processing")
	DeliveryStatusSent       = DeliveryStatus("sent")
	DeliveryStatusFailed     = DeliveryStatus("failed")
	DeliveryStatusSkipped    = DeliveryStatus("skipped")
)

func (this DeliveryStatus) String() string {
	return string(this)
}

func WrapDeliveryStatus(s string) *DeliveryStatus {
	v := DeliveryStatus(s)
	return &v
}

// Reasons a delivery was skipped rather than attempted. A skipped channel never fails the
// notification itself (BR 14).
const (
	SkipReasonChannelDisabled = "channel_disabled"
	SkipReasonExpired         = "expired"
)

const (
	DeliverySchemaName = "notification_delivery"

	DeliveryFieldId               = basemodel.FieldId
	DeliveryFieldOrgId            = basemodel.FieldOrgId
	DeliveryFieldRecipientId      = "notification_recipient_id"
	DeliveryFieldChannelName      = "channel_name"
	DeliveryFieldStatus           = "delivery_status"
	DeliveryFieldSkipReason       = "skip_reason"
	DeliveryFieldAttemptCount     = "attempt_count"
	DeliveryFieldLastAttemptAt    = "last_attempt_at"
	DeliveryFieldSentAt           = "sent_at"
	DeliveryFieldLastErrorCode    = "last_error_code"
	DeliveryFieldLastErrorMessage = "last_error_message"

	DeliveryEdgeRecipient = "recipient"
)

//go:embed delivery.json
var deliverySchemaJson string

func DeliverySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(deliverySchemaJson)
}

type Delivery struct {
	basemodel.DynamicModelBase
}

func NewDelivery() *Delivery {
	return &Delivery{basemodel.NewDynamicModel()}
}

func NewDeliveryFrom(src dmodel.DynamicFields) *Delivery {
	return &Delivery{basemodel.NewDynamicModel(src)}
}

func (this Delivery) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(DeliveryFieldOrgId)
}

func (this *Delivery) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(DeliveryFieldOrgId, v)
}

func (this Delivery) GetRecipientId() *model.Id {
	return this.GetFieldData().GetModelId(DeliveryFieldRecipientId)
}

func (this *Delivery) SetRecipientId(v *model.Id) {
	this.GetFieldData().SetModelId(DeliveryFieldRecipientId, v)
}

func (this Delivery) GetChannelName() *string {
	return this.GetFieldData().GetString(DeliveryFieldChannelName)
}

func (this *Delivery) SetChannelName(v *string) {
	this.GetFieldData().SetString(DeliveryFieldChannelName, v)
}

func (this Delivery) GetStatus() *DeliveryStatus {
	s := this.GetFieldData().GetString(DeliveryFieldStatus)
	if s == nil {
		return nil
	}
	return WrapDeliveryStatus(*s)
}

func (this *Delivery) SetStatus(v *DeliveryStatus) {
	if v == nil {
		this.GetFieldData().SetString(DeliveryFieldStatus, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(DeliveryFieldStatus, &s)
}

func (this Delivery) GetSkipReason() *string {
	return this.GetFieldData().GetString(DeliveryFieldSkipReason)
}

func (this *Delivery) SetSkipReason(v *string) {
	this.GetFieldData().SetString(DeliveryFieldSkipReason, v)
}

func (this Delivery) GetAttemptCount() *int32 {
	return this.GetFieldData().GetInt32(DeliveryFieldAttemptCount)
}

func (this *Delivery) SetAttemptCount(v *int32) {
	this.GetFieldData().SetInt32(DeliveryFieldAttemptCount, v)
}

func (this Delivery) GetLastAttemptAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(DeliveryFieldLastAttemptAt)
}

func (this *Delivery) SetLastAttemptAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(DeliveryFieldLastAttemptAt, v)
}

func (this Delivery) GetSentAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(DeliveryFieldSentAt)
}

func (this *Delivery) SetSentAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(DeliveryFieldSentAt, v)
}

func (this Delivery) GetLastErrorCode() *string {
	return this.GetFieldData().GetString(DeliveryFieldLastErrorCode)
}

func (this *Delivery) SetLastErrorCode(v *string) {
	this.GetFieldData().SetString(DeliveryFieldLastErrorCode, v)
}

func (this Delivery) GetLastErrorMessage() *string {
	return this.GetFieldData().GetString(DeliveryFieldLastErrorMessage)
}

func (this *Delivery) SetLastErrorMessage(v *string) {
	this.GetFieldData().SetString(DeliveryFieldLastErrorMessage, v)
}
