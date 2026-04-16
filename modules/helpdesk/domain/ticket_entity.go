package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketSchemaName = "helpdesk.ticket"

	TicketFieldId              = basemodel.FieldId
	TicketFieldCode            = "code"
	TicketFieldTitle           = "title"
	TicketFieldDescription     = "description"
	TicketFieldStatus          = "status"
	TicketFieldPriority        = "priority"
	TicketFieldSeverity        = "severity"
	TicketFieldSource          = "source"
	TicketFieldChannelId       = "channel_id"
	TicketFieldCategoryId      = "category_id"
	TicketFieldSlaPolicyId     = "sla_policy_id"
	TicketFieldCustomerId      = "customer_id"
	TicketFieldOrgId           = "org_id"
	TicketFieldAssignedTeamId  = "assigned_team_id"
	TicketFieldAssignedAgentId = "assigned_agent_id"
	TicketFieldProductId       = "product_id"
	TicketFieldSalesOrderId    = "sales_order_id"
	TicketFieldDueAt           = "due_at"
	TicketFieldFirstResponseAt = "first_response_at"
	TicketFieldResolvedAt      = "resolved_at"
	TicketFieldClosedAt        = "closed_at"
)

const (
	TicketEdgeActivities = "activities"
	TicketEdgeMessages   = "messages"
	TicketEdgeCategories = "categories"
)

func TicketSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket"}).
		TableName("helpdesk_tickets").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(TicketFieldCode).DataType(dmodel.FieldDataTypeString(1, 40)).RequiredForCreate().Unique()).
		Field(dmodel.DefineField().Name(TicketFieldTitle).DataType(dmodel.FieldDataTypeString(1, 255)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketFieldDescription).DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH))).
		Field(dmodel.DefineField().Name(TicketFieldStatus).DataType(dmodel.FieldDataTypeEnumString([]string{
			"new", "open", "pending_customer", "resolved", "closed", "canceled",
		})).Default("new").RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketFieldPriority).DataType(dmodel.FieldDataTypeEnumString([]string{
			"low", "medium", "high", "urgent",
		})).Default("medium").RequiredForCreate()).
		Field(dmodel.DefineField().Name(TicketFieldSeverity).DataType(dmodel.FieldDataTypeString(0, 80))).
		Field(dmodel.DefineField().Name(TicketFieldSource).DataType(dmodel.FieldDataTypeEnumString([]string{
			"email", "portal", "phone", "api", "auto",
		})).Default("portal").RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketFieldChannelId)).
		Field(basemodel.DefineFieldId(TicketFieldCategoryId)).
		Field(basemodel.DefineFieldId(TicketFieldSlaPolicyId)).
		Field(basemodel.DefineFieldId(TicketFieldCustomerId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketFieldOrgId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketFieldAssignedTeamId)).
		Field(basemodel.DefineFieldId(TicketFieldAssignedAgentId)).
		Field(basemodel.DefineFieldId(TicketFieldProductId)).
		Field(basemodel.DefineFieldId(TicketFieldSalesOrderId)).
		Field(dmodel.DefineField().Name(TicketFieldDueAt).DataType(dmodel.FieldDataTypeDateTime())).
		Field(dmodel.DefineField().Name(TicketFieldFirstResponseAt).DataType(dmodel.FieldDataTypeDateTime())).
		Field(dmodel.DefineField().Name(TicketFieldResolvedAt).DataType(dmodel.FieldDataTypeDateTime())).
		Field(dmodel.DefineField().Name(TicketFieldClosedAt).DataType(dmodel.FieldDataTypeDateTime())).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(TicketEdgeActivities).
				OneToMany(TicketActivitySchemaName, dmodel.DynamicFields{
					TicketActivityFieldTicketId: TicketFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(TicketEdgeMessages).
				OneToMany(TicketMessageSchemaName, dmodel.DynamicFields{
					TicketMessageFieldTicketId: TicketFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(TicketEdgeCategories).
				Label(model.LangJson{model.LanguageCodeEnUs: "Categories"}).
				ManyToMany(TicketCategorySchemaName, TicketCategoryRelSchemaName, "ticket").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

type Ticket struct{ basemodel.DynamicModelBase }
