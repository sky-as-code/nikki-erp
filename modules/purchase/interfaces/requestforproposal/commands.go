package requestforproposal

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateRequestForProposalCommand)(nil)
	req = (*DeleteRequestForProposalCommand)(nil)
	req = (*GetRequestForProposalQuery)(nil)
	req = (*SearchRequestForProposalsQuery)(nil)
	req = (*UpdateRequestForProposalCommand)(nil)
	req = (*SetRequestForProposalIsArchivedCommand)(nil)
	req = (*RequestForProposalExistsQuery)(nil)
	util.Unused(req)
}

var createCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "create"}

type CreateRequestForProposalCommand struct{ domain.RequestForProposal }

func (CreateRequestForProposalCommand) CqrsRequestType() cqrs.RequestType { return createCommandType }
func (CreateRequestForProposalCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RequestForProposalSchemaName)
}

type CreateRequestForProposalResult = dyn.OpResult[domain.RequestForProposal]

var updateCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "update"}

type UpdateRequestForProposalCommand struct{ domain.RequestForProposal }

func (UpdateRequestForProposalCommand) CqrsRequestType() cqrs.RequestType { return updateCommandType }
func (UpdateRequestForProposalCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.RequestForProposalSchemaName)
}

type UpdateRequestForProposalResult = dyn.OpResult[dyn.MutateResultData]

var deleteCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "delete"}

type DeleteRequestForProposalCommand dyn.DeleteOneCommand

func (DeleteRequestForProposalCommand) CqrsRequestType() cqrs.RequestType { return deleteCommandType }

type DeleteRequestForProposalResult = dyn.OpResult[dyn.MutateResultData]

var getQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "get"}

type GetRequestForProposalQuery dyn.GetOneQuery

func (GetRequestForProposalQuery) CqrsRequestType() cqrs.RequestType { return getQueryType }

type GetRequestForProposalResult = dyn.OpResult[domain.RequestForProposal]

var searchQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "search"}

type SearchRequestForProposalsQuery dyn.SearchQuery

func (SearchRequestForProposalsQuery) CqrsRequestType() cqrs.RequestType { return searchQueryType }

type SearchRequestForProposalsResultData = dyn.PagedResultData[domain.RequestForProposal]
type SearchRequestForProposalsResult = dyn.OpResult[SearchRequestForProposalsResultData]

var setArchivedCommandType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "set_archived"}

type SetRequestForProposalIsArchivedCommand dyn.SetIsArchivedCommand

func (SetRequestForProposalIsArchivedCommand) CqrsRequestType() cqrs.RequestType {
	return setArchivedCommandType
}

type SetRequestForProposalIsArchivedResult = dyn.OpResult[dyn.MutateResultData]

var existsQueryType = cqrs.RequestType{Module: "purchase", Submodule: "requestforproposal", Action: "exists"}

type RequestForProposalExistsQuery dyn.ExistsQuery

func (RequestForProposalExistsQuery) CqrsRequestType() cqrs.RequestType { return existsQueryType }

type RequestForProposalExistsResult = dyn.OpResult[dyn.ExistsResultData]
