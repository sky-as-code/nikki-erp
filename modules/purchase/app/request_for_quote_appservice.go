package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	it "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/requestforquote"
)

func NewRequestForQuoteApplicationServiceImpl(requestForQuoteSvc it.RequestForQuoteDomainService) it.RequestForQuoteAppService {
	return &RequestForQuoteApplicationServiceImpl{requestForQuoteSvc: requestForQuoteSvc}
}

type RequestForQuoteApplicationServiceImpl struct {
	requestForQuoteSvc it.RequestForQuoteDomainService
}

func (this *RequestForQuoteApplicationServiceImpl) CreateRequestForQuote(ctx corectx.Context, cmd it.CreateRequestForQuoteCommand) (*it.CreateRequestForQuoteResult, error) {
	return this.requestForQuoteSvc.CreateRequestForQuote(ctx, cmd)
}

func (this *RequestForQuoteApplicationServiceImpl) DeleteRequestForQuote(ctx corectx.Context, cmd it.DeleteRequestForQuoteCommand) (*it.DeleteRequestForQuoteResult, error) {
	return this.requestForQuoteSvc.DeleteRequestForQuote(ctx, cmd)
}

func (this *RequestForQuoteApplicationServiceImpl) RequestForQuoteExists(ctx corectx.Context, query it.RequestForQuoteExistsQuery) (*it.RequestForQuoteExistsResult, error) {
	return this.requestForQuoteSvc.RequestForQuoteExists(ctx, query)
}

func (this *RequestForQuoteApplicationServiceImpl) GetRequestForQuote(ctx corectx.Context, query it.GetRequestForQuoteQuery) (*it.GetRequestForQuoteResult, error) {
	return this.requestForQuoteSvc.GetRequestForQuote(ctx, query)
}

func (this *RequestForQuoteApplicationServiceImpl) SearchRequestForQuotes(ctx corectx.Context, query it.SearchRequestForQuotesQuery) (*it.SearchRequestForQuotesResult, error) {
	return this.requestForQuoteSvc.SearchRequestForQuotes(ctx, query)
}

func (this *RequestForQuoteApplicationServiceImpl) SetRequestForQuoteIsArchived(ctx corectx.Context, cmd it.SetRequestForQuoteIsArchivedCommand) (*it.SetRequestForQuoteIsArchivedResult, error) {
	return this.requestForQuoteSvc.SetRequestForQuoteIsArchived(ctx, cmd)
}

func (this *RequestForQuoteApplicationServiceImpl) UpdateRequestForQuote(ctx corectx.Context, cmd it.UpdateRequestForQuoteCommand) (*it.UpdateRequestForQuoteResult, error) {
	return this.requestForQuoteSvc.UpdateRequestForQuote(ctx, cmd)
}
