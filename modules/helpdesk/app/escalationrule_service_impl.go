package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

func NewEscalationRuleServiceImpl(repo it.EscalationRuleRepository, cqrsBus cqrs.CqrsBus) it.EscalationRuleService {
	return &EscalationRuleServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type EscalationRuleServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.EscalationRuleRepository
}

func (this *EscalationRuleServiceImpl) CreateEscalationRule(
	ctx corectx.Context, cmd it.CreateEscalationRuleCommand,
) (*it.CreateEscalationRuleResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.EscalationRule, *domain.EscalationRule]{Action: "create escalationRule", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *EscalationRuleServiceImpl) DeleteEscalationRule(
	ctx corectx.Context, cmd it.DeleteEscalationRuleCommand,
) (*it.DeleteEscalationRuleResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete escalationRule", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *EscalationRuleServiceImpl) GetEscalationRule(
	ctx corectx.Context, query it.GetEscalationRuleQuery,
) (*it.GetEscalationRuleResult, error) {
	return corecrud.GetOne[domain.EscalationRule](ctx, corecrud.GetOneParam{Action: "get escalationRule", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *EscalationRuleServiceImpl) EscalationRuleExists(
	ctx corectx.Context, query it.EscalationRuleExistsQuery,
) (*it.EscalationRuleExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if escalationRule exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *EscalationRuleServiceImpl) SearchEscalationRules(
	ctx corectx.Context, query it.SearchEscalationRulesQuery,
) (*it.SearchEscalationRulesResult, error) {
	return corecrud.Search[domain.EscalationRule](ctx, corecrud.SearchParam{Action: "search escalationRules", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *EscalationRuleServiceImpl) UpdateEscalationRule(
	ctx corectx.Context, cmd it.UpdateEscalationRuleCommand,
) (*it.UpdateEscalationRuleResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.EscalationRule, *domain.EscalationRule]{Action: "update escalationRule", DbRepoGetter: this.repo, Data: cmd})
}
