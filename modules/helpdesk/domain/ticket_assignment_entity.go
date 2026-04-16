package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketAssignmentSchemaName = "helpdesk.ticket_assignment"

	TicketAssignmentFieldTicketId     = "ticket_id"
	TicketAssignmentFieldAgentId      = "agent_id"
	TicketAssignmentFieldTeamId       = "team_id"
	TicketAssignmentFieldAssignedAt   = "assigned_at"
	TicketAssignmentFieldUnassignedAt = "unassigned_at"
	TicketAssignmentFieldReason       = "reason"
)

func TicketAssignmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketAssignmentSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket assignment"}).
		TableName("helpdesk_ticket_assignments").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(TicketAssignmentFieldTicketId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketAssignmentFieldAgentId)).
		Field(basemodel.DefineFieldId(TicketAssignmentFieldTeamId)).
		Field(dmodel.DefineField().Name(TicketAssignmentFieldAssignedAt).DataType(dmodel.FieldDataTypeDateTime()).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketAssignmentFieldUnassignedAt).DataType(dmodel.FieldDataTypeDateTime())).
		Field(dmodel.DefineField().Name(TicketAssignmentFieldReason).DataType(
			dmodel.FieldDataTypeEnumString([]string{"manual", "auto", "escalation"}),
		).RequiredForCreate()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TicketAssignment struct{ basemodel.DynamicModelBase }
