package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slabreach"
)

func NewSlaBreachApplicationServiceImpl(slaBreachSvc it.SlaBreachDomainService) it.SlaBreachAppService {
	return &SlaBreachApplicationServiceImpl{slaBreachSvc: slaBreachSvc}
}

type SlaBreachApplicationServiceImpl struct {
	slaBreachSvc it.SlaBreachDomainService
}

func (this *SlaBreachApplicationServiceImpl) CreateSlaBreach(ctx corectx.Context, cmd it.CreateSlaBreachCommand) (*it.CreateSlaBreachResult, error) {
	return this.slaBreachSvc.CreateSlaBreach(ctx, cmd)
}

func (this *SlaBreachApplicationServiceImpl) DeleteSlaBreach(ctx corectx.Context, cmd it.DeleteSlaBreachCommand) (*it.DeleteSlaBreachResult, error) {
	return this.slaBreachSvc.DeleteSlaBreach(ctx, cmd)
}

func (this *SlaBreachApplicationServiceImpl) GetSlaBreach(ctx corectx.Context, query it.GetSlaBreachQuery) (*it.GetSlaBreachResult, error) {
	return this.slaBreachSvc.GetSlaBreach(ctx, query)
}

func (this *SlaBreachApplicationServiceImpl) SlaBreachExists(ctx corectx.Context, query it.SlaBreachExistsQuery) (*it.SlaBreachExistsResult, error) {
	return this.slaBreachSvc.SlaBreachExists(ctx, query)
}

func (this *SlaBreachApplicationServiceImpl) SearchSlaBreaches(ctx corectx.Context, query it.SearchSlaBreachesQuery) (*it.SearchSlaBreachesResult, error) {
	return this.slaBreachSvc.SearchSlaBreaches(ctx, query)
}

func (this *SlaBreachApplicationServiceImpl) UpdateSlaBreach(ctx corectx.Context, cmd it.UpdateSlaBreachCommand) (*it.UpdateSlaBreachResult, error) {
	return this.slaBreachSvc.UpdateSlaBreach(ctx, cmd)
}
