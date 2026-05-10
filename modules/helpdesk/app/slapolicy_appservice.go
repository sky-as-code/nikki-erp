package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/helpdesk/interfaces/slapolicy"
)

func NewSlaPolicyApplicationServiceImpl(slaPolicySvc it.SlaPolicyDomainService) it.SlaPolicyAppService {
	return &SlaPolicyApplicationServiceImpl{slaPolicySvc: slaPolicySvc}
}

type SlaPolicyApplicationServiceImpl struct {
	slaPolicySvc it.SlaPolicyDomainService
}

func (this *SlaPolicyApplicationServiceImpl) CreateSlaPolicy(ctx corectx.Context, cmd it.CreateSlaPolicyCommand) (*it.CreateSlaPolicyResult, error) {
	return this.slaPolicySvc.CreateSlaPolicy(ctx, cmd)
}

func (this *SlaPolicyApplicationServiceImpl) DeleteSlaPolicy(ctx corectx.Context, cmd it.DeleteSlaPolicyCommand) (*it.DeleteSlaPolicyResult, error) {
	return this.slaPolicySvc.DeleteSlaPolicy(ctx, cmd)
}

func (this *SlaPolicyApplicationServiceImpl) GetSlaPolicy(ctx corectx.Context, query it.GetSlaPolicyQuery) (*it.GetSlaPolicyResult, error) {
	return this.slaPolicySvc.GetSlaPolicy(ctx, query)
}

func (this *SlaPolicyApplicationServiceImpl) SlaPolicyExists(ctx corectx.Context, query it.SlaPolicyExistsQuery) (*it.SlaPolicyExistsResult, error) {
	return this.slaPolicySvc.SlaPolicyExists(ctx, query)
}

func (this *SlaPolicyApplicationServiceImpl) SearchSlaPolicies(ctx corectx.Context, query it.SearchSlaPoliciesQuery) (*it.SearchSlaPoliciesResult, error) {
	return this.slaPolicySvc.SearchSlaPolicies(ctx, query)
}

func (this *SlaPolicyApplicationServiceImpl) UpdateSlaPolicy(ctx corectx.Context, cmd it.UpdateSlaPolicyCommand) (*it.UpdateSlaPolicyResult, error) {
	return this.slaPolicySvc.UpdateSlaPolicy(ctx, cmd)
}

func (this *SlaPolicyApplicationServiceImpl) SetSlaPolicyIsArchived(ctx corectx.Context, cmd it.SetSlaPolicyIsArchivedCommand) (*it.SetSlaPolicyIsArchivedResult, error) {
	return this.slaPolicySvc.SetSlaPolicyIsArchived(ctx, cmd)
}
