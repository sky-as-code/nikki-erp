package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Severity is how urgently a notification should read to the person receiving it. It carries no
// business meaning inside this module: nothing is filtered, retried or ordered by it.
type Severity string

const (
	SeverityInfo    = Severity("info")
	SeveritySuccess = Severity("success")
	SeverityWarning = Severity("warning")
	SeverityDanger  = Severity("danger")
)

func (this Severity) String() string {
	return string(this)
}

func WrapSeverity(s string) *Severity {
	v := Severity(s)
	return &v
}

// DistributionMode records whether the sender chose the channels or left them to the organization's
// configuration. Calculated at send time, never supplied by the caller (BR 11.5, BR 11.6).
type DistributionMode string

const (
	// DistributionModeAll resolves to every channel enabled for the organization.
	DistributionModeAll = DistributionMode("all")

	// DistributionModeExplicit uses exactly the channels the sender named.
	DistributionModeExplicit = DistributionMode("explicit")
)

func (this DistributionMode) String() string {
	return string(this)
}

func WrapDistributionMode(s string) *DistributionMode {
	v := DistributionMode(s)
	return &v
}

const (
	NotificationSchemaName = "notification_notification"

	NotificationFieldId                 = basemodel.FieldId
	NotificationFieldOrgId              = basemodel.FieldOrgId
	NotificationFieldSourceModule       = "source_module"
	NotificationFieldSourceResourceName = "source_resource_name"
	NotificationFieldSourceResourceKey  = "source_resource_key"
	NotificationFieldIdempotencyKey     = "idempotency_key"
	NotificationFieldTitle              = "title"
	NotificationFieldMessage            = "message"
	NotificationFieldSeverity           = "severity"
	NotificationFieldDistributionMode   = "distribution_mode"
	NotificationFieldRequestedChannels  = "requested_channels"
	NotificationFieldChannelArgs        = "channel_args"
	NotificationFieldMetadata           = "metadata"
	NotificationFieldExpiresAt          = "expires_at"
	NotificationFieldCreatedAt          = basemodel.FieldCreatedAt
)

//go:embed notification.json
var notificationSchemaJson string

func NotificationSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(notificationSchemaJson)
}

type Notification struct {
	basemodel.DynamicModelBase
}

func NewNotification() *Notification {
	return &Notification{basemodel.NewDynamicModel()}
}

func NewNotificationFrom(src dmodel.DynamicFields) *Notification {
	return &Notification{basemodel.NewDynamicModel(src)}
}

func (this Notification) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(NotificationFieldOrgId)
}

func (this *Notification) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(NotificationFieldOrgId, v)
}

func (this Notification) GetSourceModule() *string {
	return this.GetFieldData().GetString(NotificationFieldSourceModule)
}

func (this *Notification) SetSourceModule(v *string) {
	this.GetFieldData().SetString(NotificationFieldSourceModule, v)
}

func (this Notification) GetSourceResourceName() *string {
	return this.GetFieldData().GetString(NotificationFieldSourceResourceName)
}

func (this *Notification) SetSourceResourceName(v *string) {
	this.GetFieldData().SetString(NotificationFieldSourceResourceName, v)
}

func (this Notification) GetIdempotencyKey() *string {
	return this.GetFieldData().GetString(NotificationFieldIdempotencyKey)
}

func (this *Notification) SetIdempotencyKey(v *string) {
	this.GetFieldData().SetString(NotificationFieldIdempotencyKey, v)
}

func (this Notification) GetTitle() *string {
	return this.GetFieldData().GetString(NotificationFieldTitle)
}

func (this *Notification) SetTitle(v *string) {
	this.GetFieldData().SetString(NotificationFieldTitle, v)
}

func (this Notification) GetMessage() *string {
	return this.GetFieldData().GetString(NotificationFieldMessage)
}

func (this *Notification) SetMessage(v *string) {
	this.GetFieldData().SetString(NotificationFieldMessage, v)
}

func (this Notification) GetSeverity() *Severity {
	s := this.GetFieldData().GetString(NotificationFieldSeverity)
	if s == nil {
		return nil
	}
	return WrapSeverity(*s)
}

func (this *Notification) SetSeverity(v *Severity) {
	if v == nil {
		this.GetFieldData().SetString(NotificationFieldSeverity, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(NotificationFieldSeverity, &s)
}

func (this Notification) GetDistributionMode() *DistributionMode {
	s := this.GetFieldData().GetString(NotificationFieldDistributionMode)
	if s == nil {
		return nil
	}
	return WrapDistributionMode(*s)
}

func (this *Notification) SetDistributionMode(v *DistributionMode) {
	if v == nil {
		this.GetFieldData().SetString(NotificationFieldDistributionMode, nil)
		return
	}
	s := string(*v)
	this.GetFieldData().SetString(NotificationFieldDistributionMode, &s)
}

func (this Notification) GetExpiresAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(NotificationFieldExpiresAt)
}

func (this *Notification) SetExpiresAt(v *model.ModelDateTime) {
	this.GetFieldData().SetModelDateTime(NotificationFieldExpiresAt, v)
}

func (this Notification) GetCreatedAt() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(NotificationFieldCreatedAt)
}

func (this Notification) GetSourceResourceKey() map[string]any {
	return anyMap(this.GetFieldData().GetAny(NotificationFieldSourceResourceKey))
}

func (this *Notification) SetSourceResourceKey(v map[string]any) {
	this.GetFieldData().SetAny(NotificationFieldSourceResourceKey, v)
}

func (this Notification) GetRequestedChannels() map[string]any {
	return anyMap(this.GetFieldData().GetAny(NotificationFieldRequestedChannels))
}

func (this *Notification) SetRequestedChannels(v map[string]any) {
	this.GetFieldData().SetAny(NotificationFieldRequestedChannels, v)
}

func (this Notification) GetChannelArgs() map[string]any {
	return anyMap(this.GetFieldData().GetAny(NotificationFieldChannelArgs))
}

func (this *Notification) SetChannelArgs(v map[string]any) {
	this.GetFieldData().SetAny(NotificationFieldChannelArgs, v)
}

func (this Notification) GetMetadata() map[string]any {
	return anyMap(this.GetFieldData().GetAny(NotificationFieldMetadata))
}

func (this *Notification) SetMetadata(v map[string]any) {
	this.GetFieldData().SetAny(NotificationFieldMetadata, v)
}

func anyMap(raw any) map[string]any {
	if raw == nil {
		return nil
	}
	value, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return value
}
