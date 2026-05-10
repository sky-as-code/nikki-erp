package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slabreach"
)

func NewSlaBreachDomainServiceImpl(repo it.SlaBreachRepository, cqrsBus cqrs.CqrsBus) it.SlaBreachDomainService {
	return &SlaBreachDomainServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type SlaBreachDomainServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.SlaBreachRepository
}

func (this *SlaBreachDomainServiceImpl) CreateSlaBreach(
	ctx corectx.Context, cmd it.CreateSlaBreachCommand,
) (*it.CreateSlaBreachResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.SlaBreach, *models.SlaBreach]{Action: "create slaBreach", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *SlaBreachDomainServiceImpl) DeleteSlaBreach(
	ctx corectx.Context, cmd it.DeleteSlaBreachCommand,
) (*it.DeleteSlaBreachResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete slaBreach", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *SlaBreachDomainServiceImpl) GetSlaBreach(
	ctx corectx.Context, query it.GetSlaBreachQuery,
) (*it.GetSlaBreachResult, error) {
	return corecrud.GetOne[models.SlaBreach](ctx, corecrud.GetOneParam{Action: "get slaBreach", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *SlaBreachDomainServiceImpl) SlaBreachExists(
	ctx corectx.Context, query it.SlaBreachExistsQuery,
) (*it.SlaBreachExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if slaBreach exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *SlaBreachDomainServiceImpl) SearchSlaBreaches(
	ctx corectx.Context, query it.SearchSlaBreachesQuery,
) (*it.SearchSlaBreachesResult, error) {
	return corecrud.Search[models.SlaBreach](ctx, corecrud.SearchParam{Action: "search slaBreachs", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *SlaBreachDomainServiceImpl) UpdateSlaBreach(
	ctx corectx.Context, cmd it.UpdateSlaBreachCommand,
) (*it.UpdateSlaBreachResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.SlaBreach, *models.SlaBreach]{Action: "update slaBreach", DbRepoGetter: this.repo, Data: cmd})
}
