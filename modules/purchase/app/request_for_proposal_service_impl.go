package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

func NewRequestForProposalServiceImpl(repo it.RequestForProposalRepository) it.RequestForProposalService {
	return &RequestForProposalServiceImpl{repo: repo}
}

type RequestForProposalServiceImpl struct {
	repo it.RequestForProposalRepository
}

func (this *RequestForProposalServiceImpl) CreateRequestForProposal(ctx corectx.Context, cmd it.CreateRequestForProposalCommand) (*it.CreateRequestForProposalResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.RequestForProposal, *domain.RequestForProposal]{
		Action: "create request for proposal", BaseRepoGetter: this.repo, Data: cmd,
	})
}
func (this *RequestForProposalServiceImpl) DeleteRequestForProposal(ctx corectx.Context, cmd it.DeleteRequestForProposalCommand) (*it.DeleteRequestForProposalResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete request for proposal", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}
func (this *RequestForProposalServiceImpl) RequestForProposalExists(ctx corectx.Context, query it.RequestForProposalExistsQuery) (*it.RequestForProposalExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if request for proposals exist", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}
func (this *RequestForProposalServiceImpl) GetRequestForProposal(ctx corectx.Context, query it.GetRequestForProposalQuery) (*it.GetRequestForProposalResult, error) {
	return corecrud.GetOne[domain.RequestForProposal](ctx, corecrud.GetOneParam{Action: "get request for proposal", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}
func (this *RequestForProposalServiceImpl) SearchRequestForProposals(ctx corectx.Context, query it.SearchRequestForProposalsQuery) (*it.SearchRequestForProposalsResult, error) {
	return corecrud.Search[domain.RequestForProposal](ctx, corecrud.SearchParam{Action: "search request for proposals", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}
func (this *RequestForProposalServiceImpl) SetRequestForProposalIsArchived(
	ctx corectx.Context, cmd it.SetRequestForProposalIsArchivedCommand,
) (*it.SetRequestForProposalIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
func (this *RequestForProposalServiceImpl) UpdateRequestForProposal(
	ctx corectx.Context, cmd it.UpdateRequestForProposalCommand,
) (*it.UpdateRequestForProposalResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.RequestForProposal, *domain.RequestForProposal]{
		Action: "update request for proposal", DbRepoGetter: this.repo, Data: cmd,
	})
}
