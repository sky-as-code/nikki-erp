package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketActivitySchemaName = "helpdesk.ticket_activity"

	TicketActivityFieldTicketId   = "ticket_id"
	TicketActivityFieldActorId    = "actor_id"
	TicketActivityFieldType       = "type"
	TicketActivityFieldOldValue   = "old_value"
	TicketActivityFieldNewValue   = "new_value"
	TicketActivityFieldVisibility = "visibility"
)

func TicketActivitySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketActivitySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket activity"}).
		TableName("helpdesk_ticket_activities").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(TicketActivityFieldTicketId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketActivityFieldActorId)).
		Field(dmodel.DefineField().Name(TicketActivityFieldType).DataType(dmodel.FieldDataTypeEnumString([]string{
			"comment", "status_change", "assign", "escalation", "sla_breach", "field_update",
		})).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketActivityFieldOldValue).DataType(dmodel.FieldDataTypeModel())).
		Field(dmodel.DefineField().Name(TicketActivityFieldNewValue).DataType(dmodel.FieldDataTypeModel())).
		Field(dmodel.DefineField().Name(TicketActivityFieldVisibility).DataType(dmodel.FieldDataTypeEnumString([]string{
			"internal", "customer",
		})).Default("internal")).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TicketActivity struct{ basemodel.DynamicModelBase }
