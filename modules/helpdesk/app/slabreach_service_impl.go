package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slabreach"
)

func NewSlaBreachServiceImpl(repo it.SlaBreachRepository, cqrsBus cqrs.CqrsBus) it.SlaBreachService {
	return &SlaBreachServiceImpl{cqrsBus: cqrsBus, repo: repo}
}

type SlaBreachServiceImpl struct {
	cqrsBus cqrs.CqrsBus
	repo    it.SlaBreachRepository
}

func (this *SlaBreachServiceImpl) CreateSlaBreach(
	ctx corectx.Context, cmd it.CreateSlaBreachCommand,
) (*it.CreateSlaBreachResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.SlaBreach, *domain.SlaBreach]{Action: "create slaBreach", BaseRepoGetter: this.repo, Data: cmd})
}

func (this *SlaBreachServiceImpl) DeleteSlaBreach(
	ctx corectx.Context, cmd it.DeleteSlaBreachCommand,
) (*it.DeleteSlaBreachResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete slaBreach", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}

func (this *SlaBreachServiceImpl) GetSlaBreach(
	ctx corectx.Context, query it.GetSlaBreachQuery,
) (*it.GetSlaBreachResult, error) {
	return corecrud.GetOne[domain.SlaBreach](ctx, corecrud.GetOneParam{Action: "get slaBreach", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}

func (this *SlaBreachServiceImpl) SlaBreachExists(
	ctx corectx.Context, query it.SlaBreachExistsQuery,
) (*it.SlaBreachExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if slaBreach exists", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}

func (this *SlaBreachServiceImpl) SearchSlaBreaches(
	ctx corectx.Context, query it.SearchSlaBreachesQuery,
) (*it.SearchSlaBreachesResult, error) {
	return corecrud.Search[domain.SlaBreach](ctx, corecrud.SearchParam{Action: "search slaBreachs", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}

func (this *SlaBreachServiceImpl) UpdateSlaBreach(
	ctx corectx.Context, cmd it.UpdateSlaBreachCommand,
) (*it.UpdateSlaBreachResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.SlaBreach, *domain.SlaBreach]{Action: "update slaBreach", DbRepoGetter: this.repo, Data: cmd})
}
