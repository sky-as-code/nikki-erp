package services

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforproposal"
)

func NewRequestForProposalDomainServiceImpl(repo it.RequestForProposalRepository) it.RequestForProposalDomainService {
	return &RequestForProposalDomainServiceImpl{repo: repo}
}

type RequestForProposalDomainServiceImpl struct {
	repo it.RequestForProposalRepository
}

func (this *RequestForProposalDomainServiceImpl) CreateRequestForProposal(ctx corectx.Context, cmd it.CreateRequestForProposalCommand) (*it.CreateRequestForProposalResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.RequestForProposal, *models.RequestForProposal]{
		Action: "create request for proposal", BaseRepoGetter: this.repo, Data: cmd,
	})
}
func (this *RequestForProposalDomainServiceImpl) DeleteRequestForProposal(ctx corectx.Context, cmd it.DeleteRequestForProposalCommand) (*it.DeleteRequestForProposalResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete request for proposal", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}
func (this *RequestForProposalDomainServiceImpl) RequestForProposalExists(ctx corectx.Context, query it.RequestForProposalExistsQuery) (*it.RequestForProposalExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if request for proposals exist", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}
func (this *RequestForProposalDomainServiceImpl) GetRequestForProposal(ctx corectx.Context, query it.GetRequestForProposalQuery) (*it.GetRequestForProposalResult, error) {
	return corecrud.GetOne[models.RequestForProposal](ctx, corecrud.GetOneParam{Action: "get request for proposal", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}
func (this *RequestForProposalDomainServiceImpl) SearchRequestForProposals(ctx corectx.Context, query it.SearchRequestForProposalsQuery) (*it.SearchRequestForProposalsResult, error) {
	return corecrud.Search[models.RequestForProposal](ctx, corecrud.SearchParam{Action: "search request for proposals", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}
func (this *RequestForProposalDomainServiceImpl) SetRequestForProposalIsArchived(
	ctx corectx.Context, cmd it.SetRequestForProposalIsArchivedCommand,
) (*it.SetRequestForProposalIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
func (this *RequestForProposalDomainServiceImpl) UpdateRequestForProposal(
	ctx corectx.Context, cmd it.UpdateRequestForProposalCommand,
) (*it.UpdateRequestForProposalResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.RequestForProposal, *models.RequestForProposal]{
		Action: "update request for proposal", DbRepoGetter: this.repo, Data: cmd,
	})
}
