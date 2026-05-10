package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

func NewRequestForProposalApplicationServiceImpl(requestForProposalSvc it.RequestForProposalDomainService) it.RequestForProposalAppService {
	return &RequestForProposalApplicationServiceImpl{requestForProposalSvc: requestForProposalSvc}
}

type RequestForProposalApplicationServiceImpl struct {
	requestForProposalSvc it.RequestForProposalDomainService
}

func (this *RequestForProposalApplicationServiceImpl) CreateRequestForProposal(ctx corectx.Context, cmd it.CreateRequestForProposalCommand) (*it.CreateRequestForProposalResult, error) {
	return this.requestForProposalSvc.CreateRequestForProposal(ctx, cmd)
}

func (this *RequestForProposalApplicationServiceImpl) DeleteRequestForProposal(ctx corectx.Context, cmd it.DeleteRequestForProposalCommand) (*it.DeleteRequestForProposalResult, error) {
	return this.requestForProposalSvc.DeleteRequestForProposal(ctx, cmd)
}

func (this *RequestForProposalApplicationServiceImpl) RequestForProposalExists(ctx corectx.Context, query it.RequestForProposalExistsQuery) (*it.RequestForProposalExistsResult, error) {
	return this.requestForProposalSvc.RequestForProposalExists(ctx, query)
}

func (this *RequestForProposalApplicationServiceImpl) GetRequestForProposal(ctx corectx.Context, query it.GetRequestForProposalQuery) (*it.GetRequestForProposalResult, error) {
	return this.requestForProposalSvc.GetRequestForProposal(ctx, query)
}

func (this *RequestForProposalApplicationServiceImpl) SearchRequestForProposals(ctx corectx.Context, query it.SearchRequestForProposalsQuery) (*it.SearchRequestForProposalsResult, error) {
	return this.requestForProposalSvc.SearchRequestForProposals(ctx, query)
}

func (this *RequestForProposalApplicationServiceImpl) SetRequestForProposalIsArchived(ctx corectx.Context, cmd it.SetRequestForProposalIsArchivedCommand) (*it.SetRequestForProposalIsArchivedResult, error) {
	return this.requestForProposalSvc.SetRequestForProposalIsArchived(ctx, cmd)
}

func (this *RequestForProposalApplicationServiceImpl) UpdateRequestForProposal(ctx corectx.Context, cmd it.UpdateRequestForProposalCommand) (*it.UpdateRequestForProposalResult, error) {
	return this.requestForProposalSvc.UpdateRequestForProposal(ctx, cmd)
}
