package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

func NewSlaPolicyServiceImpl(repo it.SlaPolicyRepository, cqrsBus cqrs.CqrsBus) it.SlaPolicyService {
	return &SlaPolicyServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type SlaPolicyServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.SlaPolicyRepository
}

func (this *SlaPolicyServiceImpl) CreateSlaPolicy(
	ctx corectx.Context, cmd it.CreateSlaPolicyCommand,
) (*it.CreateSlaPolicyResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.SlaPolicy, *domain.SlaPolicy]{Action: "create slaPolicy", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *SlaPolicyServiceImpl) DeleteSlaPolicy(
	ctx corectx.Context, cmd it.DeleteSlaPolicyCommand,
) (*it.DeleteSlaPolicyResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete slaPolicy", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *SlaPolicyServiceImpl) GetSlaPolicy(
	ctx corectx.Context, query it.GetSlaPolicyQuery,
) (*it.GetSlaPolicyResult, error) {
	return corecrud.GetOne[domain.SlaPolicy](ctx, corecrud.GetOneParam{Action: "get slaPolicy", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *SlaPolicyServiceImpl) SlaPolicyExists(
	ctx corectx.Context, query it.SlaPolicyExistsQuery,
) (*it.SlaPolicyExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if slaPolicy exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *SlaPolicyServiceImpl) SearchSlaPolicies(
	ctx corectx.Context, query it.SearchSlaPoliciesQuery,
) (*it.SearchSlaPoliciesResult, error) {
	return corecrud.Search[domain.SlaPolicy](ctx, corecrud.SearchParam{Action: "search slaPolicys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *SlaPolicyServiceImpl) UpdateSlaPolicy(
	ctx corectx.Context, cmd it.UpdateSlaPolicyCommand,
) (*it.UpdateSlaPolicyResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.SlaPolicy, *domain.SlaPolicy]{Action: "update slaPolicy", DbRepoGetter: this.repo, Data: cmd})
}

func (this *SlaPolicyServiceImpl) SetSlaPolicyIsArchived(
	ctx corectx.Context, cmd it.SetSlaPolicyIsArchivedCommand,
) (*it.SetSlaPolicyIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
