package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	AuditEventSchemaName = "purchase_audit_event"

	AuditEventFieldId         = basemodel.FieldId
	AuditEventFieldOrgId      = basemodel.FieldOrgId
	AuditEventFieldEntityType = "entity_type"
	AuditEventFieldEntityId   = "entity_id"
	AuditEventFieldAction     = "action"
	AuditEventFieldActorId    = "actor_id"
	AuditEventFieldFromStatus = "from_status"
	AuditEventFieldToStatus   = "to_status"
	AuditEventFieldReason     = "reason"
	AuditEventFieldMetadata   = "metadata"
)

//go:embed audit_event.json
var auditEventSchemaJson string

func AuditEventSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(auditEventSchemaJson)
}

type AuditEvent struct {
	fields dmodel.DynamicFields
}

func NewAuditEvent() *AuditEvent {
	return &AuditEvent{fields: make(dmodel.DynamicFields)}
}

func NewAuditEventFrom(src dmodel.DynamicFields) *AuditEvent {
	return &AuditEvent{fields: src}
}

func (this AuditEvent) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *AuditEvent) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this AuditEvent) GetId() *model.Id {
	return this.fields.GetModelId(AuditEventFieldId)
}

func (this *AuditEvent) SetId(v *model.Id) {
	this.fields.SetModelId(AuditEventFieldId, v)
}

func (this AuditEvent) GetOrgId() *model.Id {
	return this.fields.GetModelId(AuditEventFieldOrgId)
}

func (this *AuditEvent) SetOrgId(v *model.Id) {
	this.fields.SetModelId(AuditEventFieldOrgId, v)
}

func (this AuditEvent) GetEntityType() *string {
	return this.fields.GetString(AuditEventFieldEntityType)
}

func (this *AuditEvent) SetEntityType(v *string) {
	this.fields.SetString(AuditEventFieldEntityType, v)
}

func (this AuditEvent) GetEntityId() *model.Id {
	return this.fields.GetModelId(AuditEventFieldEntityId)
}

func (this *AuditEvent) SetEntityId(v *model.Id) {
	this.fields.SetModelId(AuditEventFieldEntityId, v)
}

func (this AuditEvent) GetAction() *string {
	return this.fields.GetString(AuditEventFieldAction)
}

func (this *AuditEvent) SetAction(v *string) {
	this.fields.SetString(AuditEventFieldAction, v)
}

func (this AuditEvent) GetActorId() *model.Id {
	return this.fields.GetModelId(AuditEventFieldActorId)
}

func (this *AuditEvent) SetActorId(v *model.Id) {
	this.fields.SetModelId(AuditEventFieldActorId, v)
}

func (this AuditEvent) GetFromStatus() *string {
	return this.fields.GetString(AuditEventFieldFromStatus)
}

func (this *AuditEvent) SetFromStatus(v *string) {
	this.fields.SetString(AuditEventFieldFromStatus, v)
}

func (this AuditEvent) GetToStatus() *string {
	return this.fields.GetString(AuditEventFieldToStatus)
}

func (this *AuditEvent) SetToStatus(v *string) {
	this.fields.SetString(AuditEventFieldToStatus, v)
}

func (this AuditEvent) GetReason() *string {
	return this.fields.GetString(AuditEventFieldReason)
}

func (this *AuditEvent) SetReason(v *string) {
	this.fields.SetString(AuditEventFieldReason, v)
}

func (this AuditEvent) GetMetadata() any {
	return this.fields.GetAny(AuditEventFieldMetadata)
}

func (this *AuditEvent) SetMetadata(v any) {
	this.fields.SetAny(AuditEventFieldMetadata, v)
}
