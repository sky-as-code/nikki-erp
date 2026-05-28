package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

func NewSlaPolicyDomainServiceImpl(repo it.SlaPolicyRepository, cqrsBus cqrs.CqrsBus) it.SlaPolicyDomainService {
	return &SlaPolicyDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type SlaPolicyDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.SlaPolicyRepository
}

func (this *SlaPolicyDomainServiceImpl) CreateSlaPolicy(
	ctx corectx.Context, cmd it.CreateSlaPolicyCommand,
) (*it.CreateSlaPolicyResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.SlaPolicy, *models.SlaPolicy]{Action: "create slaPolicy", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *SlaPolicyDomainServiceImpl) DeleteSlaPolicy(
	ctx corectx.Context, cmd it.DeleteSlaPolicyCommand,
) (*it.DeleteSlaPolicyResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete slaPolicy", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *SlaPolicyDomainServiceImpl) GetSlaPolicy(
	ctx corectx.Context, query it.GetSlaPolicyQuery,
) (*it.GetSlaPolicyResult, error) {
	return corecrud.GetOne[models.SlaPolicy](ctx, corecrud.GetOneParam{Action: "get slaPolicy", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *SlaPolicyDomainServiceImpl) SlaPolicyExists(
	ctx corectx.Context, query it.SlaPolicyExistsQuery,
) (*it.SlaPolicyExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if slaPolicy exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *SlaPolicyDomainServiceImpl) SearchSlaPolicies(
	ctx corectx.Context, query it.SearchSlaPoliciesQuery,
) (*it.SearchSlaPoliciesResult, error) {
	return corecrud.Search[models.SlaPolicy](ctx, corecrud.SearchParam{Action: "search slaPolicys", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *SlaPolicyDomainServiceImpl) UpdateSlaPolicy(
	ctx corectx.Context, cmd it.UpdateSlaPolicyCommand,
) (*it.UpdateSlaPolicyResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.SlaPolicy, *models.SlaPolicy]{Action: "update slaPolicy", DbRepoGetter: this.repo, Data: cmd})
}

func (this *SlaPolicyDomainServiceImpl) SetSlaPolicyIsArchived(
	ctx corectx.Context, cmd it.SetSlaPolicyIsArchivedCommand,
) (*it.SetSlaPolicyIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
