package escalationrule

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type EscalationRuleService interface {
	CreateEscalationRule(ctx corectx.Context, cmd CreateEscalationRuleCommand) (*CreateEscalationRuleResult, error)
	DeleteEscalationRule(ctx corectx.Context, cmd DeleteEscalationRuleCommand) (*DeleteEscalationRuleResult, error)
	GetEscalationRule(ctx corectx.Context, query GetEscalationRuleQuery) (*GetEscalationRuleResult, error)
	EscalationRuleExists(ctx corectx.Context, query EscalationRuleExistsQuery) (*EscalationRuleExistsResult, error)
	SearchEscalationRules(ctx corectx.Context, query SearchEscalationRulesQuery) (*SearchEscalationRulesResult, error)
	UpdateEscalationRule(ctx corectx.Context, cmd UpdateEscalationRuleCommand) (*UpdateEscalationRuleResult, error)
}
