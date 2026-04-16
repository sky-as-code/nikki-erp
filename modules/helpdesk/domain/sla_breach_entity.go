package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SlaBreachSchemaName = "helpdesk.sla_breach"

	SlaBreachFieldTicketId    = "ticket_id"
	SlaBreachFieldSlaPolicyId = "sla_policy_id"
	SlaBreachFieldBreachType  = "breach_type"
	SlaBreachFieldBreachedAt  = "breached_at"
)

func SlaBreachSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(SlaBreachSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Sla breach"}).
		TableName("helpdesk_sla_breaches").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(SlaBreachFieldTicketId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(SlaBreachFieldSlaPolicyId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(SlaBreachFieldBreachType).DataType(dmodel.FieldDataTypeEnumString([]string{
			"response", "resolution",
		})).RequiredForCreate()).
		Field(dmodel.DefineField().Name(SlaBreachFieldBreachedAt).DataType(dmodel.FieldDataTypeDateTime()).RequiredForCreate()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type SlaBreach struct{ basemodel.DynamicModelBase }
