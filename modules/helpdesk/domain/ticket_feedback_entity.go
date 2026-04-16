package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketFeedbackSchemaName = "helpdesk.ticket_feedback"

	TicketFeedbackFieldTicketId = "ticket_id"
	TicketFeedbackFieldRating   = "rating"
	TicketFeedbackFieldComment  = "comment"
)

func TicketFeedbackSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketFeedbackSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket feedback"}).
		TableName("helpdesk_ticket_feedbacks").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(TicketFeedbackFieldTicketId).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(TicketFeedbackFieldRating).DataType(dmodel.FieldDataTypeInt32(1, 5)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketFeedbackFieldComment).DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH))).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TicketFeedback struct{ basemodel.DynamicModelBase }
