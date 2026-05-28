package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

func NewEscalationRuleApplicationServiceImpl(escalationRuleSvc it.EscalationRuleDomainService) it.EscalationRuleAppService {
	return &EscalationRuleApplicationServiceImpl{escalationRuleSvc: escalationRuleSvc}
}

type EscalationRuleApplicationServiceImpl struct {
	escalationRuleSvc it.EscalationRuleDomainService
}

func (this *EscalationRuleApplicationServiceImpl) CreateEscalationRule(ctx corectx.Context, cmd it.CreateEscalationRuleCommand) (*it.CreateEscalationRuleResult, error) {
	return this.escalationRuleSvc.CreateEscalationRule(ctx, cmd)
}

func (this *EscalationRuleApplicationServiceImpl) DeleteEscalationRule(ctx corectx.Context, cmd it.DeleteEscalationRuleCommand) (*it.DeleteEscalationRuleResult, error) {
	return this.escalationRuleSvc.DeleteEscalationRule(ctx, cmd)
}

func (this *EscalationRuleApplicationServiceImpl) GetEscalationRule(ctx corectx.Context, query it.GetEscalationRuleQuery) (*it.GetEscalationRuleResult, error) {
	return this.escalationRuleSvc.GetEscalationRule(ctx, query)
}

func (this *EscalationRuleApplicationServiceImpl) EscalationRuleExists(ctx corectx.Context, query it.EscalationRuleExistsQuery) (*it.EscalationRuleExistsResult, error) {
	return this.escalationRuleSvc.EscalationRuleExists(ctx, query)
}

func (this *EscalationRuleApplicationServiceImpl) SearchEscalationRules(ctx corectx.Context, query it.SearchEscalationRulesQuery) (*it.SearchEscalationRulesResult, error) {
	return this.escalationRuleSvc.SearchEscalationRules(ctx, query)
}

func (this *EscalationRuleApplicationServiceImpl) UpdateEscalationRule(ctx corectx.Context, cmd it.UpdateEscalationRuleCommand) (*it.UpdateEscalationRuleResult, error) {
	return this.escalationRuleSvc.UpdateEscalationRule(ctx, cmd)
}
