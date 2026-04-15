package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforquote"
)

func NewRequestForQuoteServiceImpl(repo it.RequestForQuoteRepository) it.RequestForQuoteService {
	return &RequestForQuoteServiceImpl{repo: repo}
}

type RequestForQuoteServiceImpl struct{ repo it.RequestForQuoteRepository }

func (this *RequestForQuoteServiceImpl) CreateRequestForQuote(ctx corectx.Context, cmd it.CreateRequestForQuoteCommand) (*it.CreateRequestForQuoteResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[domain.RequestForQuote, *domain.RequestForQuote]{
		Action: "create request for quote", BaseRepoGetter: this.repo, Data: cmd,
	})
}
func (this *RequestForQuoteServiceImpl) DeleteRequestForQuote(ctx corectx.Context, cmd it.DeleteRequestForQuoteCommand) (*it.DeleteRequestForQuoteResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{Action: "delete request for quote", DbRepoGetter: this.repo, Cmd: dyn.DeleteOneCommand(cmd)})
}
func (this *RequestForQuoteServiceImpl) RequestForQuoteExists(ctx corectx.Context, query it.RequestForQuoteExistsQuery) (*it.RequestForQuoteExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{Action: "check if request for quotes exist", DbRepoGetter: this.repo, Query: dyn.ExistsQuery(query)})
}
func (this *RequestForQuoteServiceImpl) GetRequestForQuote(ctx corectx.Context, query it.GetRequestForQuoteQuery) (*it.GetRequestForQuoteResult, error) {
	return corecrud.GetOne[domain.RequestForQuote](ctx, corecrud.GetOneParam{Action: "get request for quote", DbRepoGetter: this.repo, Query: dyn.GetOneQuery(query)})
}
func (this *RequestForQuoteServiceImpl) SearchRequestForQuotes(ctx corectx.Context, query it.SearchRequestForQuotesQuery) (*it.SearchRequestForQuotesResult, error) {
	return corecrud.Search[domain.RequestForQuote](ctx, corecrud.SearchParam{Action: "search request for quotes", DbRepoGetter: this.repo, Query: dyn.SearchQuery(query)})
}
func (this *RequestForQuoteServiceImpl) SetRequestForQuoteIsArchived(
	ctx corectx.Context, cmd it.SetRequestForQuoteIsArchivedCommand,
) (*it.SetRequestForQuoteIsArchivedResult, error) {
	return corecrud.SetIsArchived(ctx, this.repo, dyn.SetIsArchivedCommand(cmd))
}
func (this *RequestForQuoteServiceImpl) UpdateRequestForQuote(ctx corectx.Context, cmd it.UpdateRequestForQuoteCommand) (*it.UpdateRequestForQuoteResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[domain.RequestForQuote, *domain.RequestForQuote]{
		Action: "update request for quote", DbRepoGetter: this.repo, Data: cmd,
	})
}
