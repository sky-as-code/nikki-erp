package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RecipientSchemaName = "notification_recipient"

	RecipientFieldId              = basemodel.FieldId
	RecipientFieldOrgId           = basemodel.FieldOrgId
	RecipientFieldNotificationId  = "notification_id"
	RecipientFieldRecipientUserId = "recipient_user_id"
	RecipientFieldStreamSeq       = "stream_seq"
	RecipientFieldReadAt          = "read_at"
	RecipientFieldCreatedAt       = basemodel.FieldCreatedAt

	RecipientEdgeNotification = "notification"
)

//go:embed recipient.json
var recipientSchemaJson string

func RecipientSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(recipientSchemaJson)
}

type Recipient struct {
	basemodel.DynamicModelBase
}

func NewRecipient() *Recipient {
	return &Recipient{basemodel.NewDynamicModel()}
}

func NewRecipientFrom(src dmodel.DynamicFields) *Recipient {
	return &Recipient{basemodel.NewDynamicModel(src)}
}

func (this Recipient) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(RecipientFieldOrgId)
}

func (this *Recipient) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(RecipientFieldOrgId, v)
}

func (this Recipient) GetNotificationId() *model.Id {
	return this.GetFieldData().GetModelId(RecipientFieldNotificationId)
}

func (this *Recipient) SetNotificationId(v *model.Id) {
	this.GetFieldData().SetModelId(RecipientFieldNotificationId, v)
}

func (this Recipient) GetRecipientUserId() *model.Id {
	return this.GetFieldData().GetModelId(RecipientFieldRecipientUserId)
}

func (this *Recipient) SetRecipientUserId(v *model.Id) {
	this.GetFieldData().SetModelId(RecipientFieldRecipientUserId, v)
}

func (this Recipient) GetStreamSeq() *int64 {
	return this.GetFieldData().GetInt64(RecipientFieldStreamSeq)
}

func (this *Recipient) SetStreamSeq(v *int64) {
	this.GetFieldData().SetInt64(RecipientFieldStreamSeq, v)
}

func (this Recipient) GetReadAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(RecipientFieldReadAt)
}

func (this *Recipient) SetReadAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(RecipientFieldReadAt, v)
}

func (this Recipient) GetCreatedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(RecipientFieldCreatedAt)
}

// IsRead is calculated, never stored: read state is exactly "read_at is set" (BR 7), and a
// persisted duplicate could disagree with it.
func (this Recipient) IsRead() bool {
	return this.GetReadAt() != nil
}
