package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/escalationrule"
)

func NewEscalationRuleDomainServiceImpl(repo it.EscalationRuleRepository, cqrsBus cqrs.CqrsBus) it.EscalationRuleDomainService {
	return &EscalationRuleDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type EscalationRuleDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.EscalationRuleRepository
}

func (this *EscalationRuleDomainServiceImpl) CreateEscalationRule(
	ctx corectx.Context, cmd it.CreateEscalationRuleCommand,
) (*it.CreateEscalationRuleResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.EscalationRule, *models.EscalationRule]{Action: "create escalationRule", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *EscalationRuleDomainServiceImpl) DeleteEscalationRule(
	ctx corectx.Context, cmd it.DeleteEscalationRuleCommand,
) (*it.DeleteEscalationRuleResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete escalationRule", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *EscalationRuleDomainServiceImpl) GetEscalationRule(
	ctx corectx.Context, query it.GetEscalationRuleQuery,
) (*it.GetEscalationRuleResult, error) {
	return corecrud.GetOne[models.EscalationRule](ctx, corecrud.GetOneParam{Action: "get escalationRule", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *EscalationRuleDomainServiceImpl) EscalationRuleExists(
	ctx corectx.Context, query it.EscalationRuleExistsQuery,
) (*it.EscalationRuleExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if escalationRule exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *EscalationRuleDomainServiceImpl) SearchEscalationRules(
	ctx corectx.Context, query it.SearchEscalationRulesQuery,
) (*it.SearchEscalationRulesResult, error) {
	return corecrud.Search[models.EscalationRule](ctx, corecrud.SearchParam{Action: "search escalationRules", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *EscalationRuleDomainServiceImpl) UpdateEscalationRule(
	ctx corectx.Context, cmd it.UpdateEscalationRuleCommand,
) (*it.UpdateEscalationRuleResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.EscalationRule, *models.EscalationRule]{Action: "update escalationRule", DbRepoGetter: this.repo, Data: cmd})
}
