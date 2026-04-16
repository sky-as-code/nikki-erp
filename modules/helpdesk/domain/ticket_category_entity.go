package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TicketCategorySchemaName    = "helpdesk.ticket_category"
	TicketCategoryRelSchemaName = "helpdesk.ticket_category_rel"

	TicketCategoryFieldName               = "name"
	TicketCategoryFieldParentId           = "parent_id"
	TicketCategoryFieldDefaultSlaPolicyId = "default_sla_policy_id"
	TicketCategoryFieldDefaultTeamId      = "default_team_id"

	TicketCategoryEdgeTickets = "tickets"
)

func TicketCategorySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketCategorySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Ticket category"}).
		TableName("helpdesk_ticket_categories").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(TicketCategoryFieldName).DataType(dmodel.FieldDataTypeString(1, 120)).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TicketCategoryFieldParentId)).
		Field(basemodel.DefineFieldId(TicketCategoryFieldDefaultSlaPolicyId)).
		Field(basemodel.DefineFieldId(TicketCategoryFieldDefaultTeamId)).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(TicketCategoryEdgeTickets).
				Label(model.LangJson{model.LanguageCodeEnUs: "Tickets"}).
				ManyToMany(TicketSchemaName, TicketCategoryRelSchemaName, "ticket_category").
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

func TicketCategoryRelSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TicketCategoryRelSchemaName).
		TableName("helpdesk_ticket_category_rel").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique("ticket_id", "ticket_category_id").
		Field(basemodel.DefineFieldId("ticket_id").RequiredForCreate()).
		Field(basemodel.DefineFieldId("ticket_category_id").RequiredForCreate())
}

type TicketCategory struct{ basemodel.DynamicModelBase }
