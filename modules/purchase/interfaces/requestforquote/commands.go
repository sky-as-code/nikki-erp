package requestforquote

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateRequestForQuoteCommand)(nil)
	req = (*DeleteRequestForQuoteCommand)(nil)
	req = (*GetRequestForQuoteQuery)(nil)
	req = (*SearchRequestForQuotesQuery)(nil)
	req = (*UpdateRequestForQuoteCommand)(nil)
	req = (*SetRequestForQuoteIsArchivedCommand)(nil)
	req = (*RequestForQuoteExistsQuery)(nil)
	util.Unused(req)
}

var createCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "create"}

type CreateRequestForQuoteCommand struct{ domain.RequestForQuote }

func (CreateRequestForQuoteCommand) CqrsRequestType() cqrs.RequestType { return createCommandType }
func (CreateRequestForQuoteCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RequestForQuoteSchemaName)
}

type CreateRequestForQuoteResult = dyn.OpResult[domain.RequestForQuote]

var updateCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "update"}

type UpdateRequestForQuoteCommand struct{ domain.RequestForQuote }

func (UpdateRequestForQuoteCommand) CqrsRequestType() cqrs.RequestType { return updateCommandType }
func (UpdateRequestForQuoteCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RequestForQuoteSchemaName)
}

type UpdateRequestForQuoteResult = dyn.OpResult[dyn.MutateResultData]

var deleteCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "delete"}

type DeleteRequestForQuoteCommand dyn.DeleteOneCommand

func (DeleteRequestForQuoteCommand) CqrsRequestType() cqrs.RequestType { return deleteCommandType }

type DeleteRequestForQuoteResult = dyn.OpResult[dyn.MutateResultData]

var getQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "get"}

type GetRequestForQuoteQuery dyn.GetOneQuery

func (GetRequestForQuoteQuery) CqrsRequestType() cqrs.RequestType { return getQueryType }

type GetRequestForQuoteResult = dyn.OpResult[domain.RequestForQuote]

var searchQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "search"}

type SearchRequestForQuotesQuery dyn.SearchQuery

func (SearchRequestForQuotesQuery) CqrsRequestType() cqrs.RequestType { return searchQueryType }

type SearchRequestForQuotesResultData = dyn.PagedResultData[domain.RequestForQuote]
type SearchRequestForQuotesResult = dyn.OpResult[SearchRequestForQuotesResultData]

var setArchivedCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "set_archived"}

type SetRequestForQuoteIsArchivedCommand dyn.SetIsArchivedCommand

func (SetRequestForQuoteIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setArchivedCommandType
}

type SetRequestForQuoteIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var existsQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforquote", Action: "exists"}

type RequestForQuoteExistsQuery dyn.ExistsQuery

func (RequestForQuoteExistsQuery) CqrsRequestType() cqrs.RequestType { return existsQueryType }

type RequestForQuoteExistsResult = dyn.OpResult[dyn.ExistsResultData]
