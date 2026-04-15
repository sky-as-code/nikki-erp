package requestforquote

import corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

type RequestForQuoteService interface {
	CreateRequestForQuote(ctx corectx.Context, cmd CreateRequestForQuoteCommand) (*CreateRequestForQuoteResult, error)
	DeleteRequestForQuote(ctx corectx.Context, cmd DeleteRequestForQuoteCommand) (*DeleteRequestForQuoteResult, error)
	RequestForQuoteExists(ctx corectx.Context, query RequestForQuoteExistsQuery) (*RequestForQuoteExistsResult, error)
	GetRequestForQuote(ctx corectx.Context, query GetRequestForQuoteQuery) (*GetRequestForQuoteResult, error)
	SearchRequestForQuotes(ctx corectx.Context, query SearchRequestForQuotesQuery) (*SearchRequestForQuotesResult, error)
	SetRequestForQuoteIsArchived(ctx corectx.Context, cmd SetRequestForQuoteIsArchivedCommand) (*SetRequestForQuoteIsArchivedResult, error)
	UpdateRequestForQuote(ctx corectx.Context, cmd UpdateRequestForQuoteCommand) (*UpdateRequestForQuoteResult, error)
}
