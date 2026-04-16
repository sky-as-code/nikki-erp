package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	EscalationRuleSchemaName = "helpdesk.escalation_rule"

	EscalationRuleFieldSlaPolicyId      = "sla_policy_id"
	EscalationRuleFieldAfterMinutes     = "after_minutes"
	EscalationRuleFieldEscalateToTeamId = "escalate_to_team_id"
	EscalationRuleFieldEscalateToUserId = "escalate_to_user_id"
	EscalationRuleFieldPriorityUpgrade  = "priority_upgrade"
)

func EscalationRuleSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(EscalationRuleSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Escalation rule"}).
		TableName("helpdesk_escalation_rules").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(basemodel.DefineFieldId(EscalationRuleFieldSlaPolicyId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(EscalationRuleFieldAfterMinutes).DataType(dmodel.FieldDataTypeInt32(0, 1000000)).RequiredForCreate()).
		Field(basemodel.DefineFieldId(EscalationRuleFieldEscalateToTeamId)).
		Field(basemodel.DefineFieldId(EscalationRuleFieldEscalateToUserId)).
		Field(dmodel.DefineField().Name(EscalationRuleFieldPriorityUpgrade).DataType(dmodel.FieldDataTypeEnumString([]string{
			"low", "medium", "high", "urgent",
		}))).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type EscalationRule struct{ basemodel.DynamicModelBase }
