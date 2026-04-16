package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketMessageSchemaName = "helpdesk.ticket_message"

	TicketMessageFieldTicketId       = "ticket_id"
	TicketMessageFieldSenderType     = "sender_type"
	TicketMessageFieldSenderId       = "sender_id"
	TicketMessageFieldBody           = "body"
	TicketMessageFieldAttachments    = "attachments"
	TicketMessageFieldIsInternalNote = "is_internal_note"
)

func TicketMessageSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketMessageSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket message"}).
		TableName("helpdesk_ticket_messages").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(TicketMessageFieldTicketId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketMessageFieldSenderType).DataType(dmodel.FieldDataTypeEnumString([]string{
			"agent", "customer", "system",
		})).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketMessageFieldSenderId)).
		Field(dmodel.DefineField().Name(TicketMessageFieldBody).DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH))).
		Field(dmodel.DefineField().Name(TicketMessageFieldAttachments).DataType(dmodel.FieldDataTypeModel())).
		Field(dmodel.DefineField().Name(TicketMessageFieldIsInternalNote).DataType(dmodel.FieldDataTypeBoolean()).Default(false)).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TicketMessage struct{ basemodel.DynamicModelBase }
