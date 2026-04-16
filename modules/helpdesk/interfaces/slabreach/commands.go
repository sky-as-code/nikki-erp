package slabreach

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/helpdesk/domain"
)

func init() {
	var req cqrs.Request
	req = (*CreateSlaBreachCommand)(nil)
	req = (*DeleteSlaBreachCommand)(nil)
	req = (*GetSlaBreachQuery)(nil)
	req = (*SlaBreachExistsQuery)(nil)
	req = (*SearchSlaBreachesQuery)(nil)
	req = (*UpdateSlaBreachCommand)(nil)
	util.Unused(req)
}

var createSlaBreachCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "createSlaBreach"}

type CreateSlaBreachCommand struct{ domain.SlaBreach }

func (CreateSlaBreachCommand) CqrsRequestType() cqrs.RequestType { return createSlaBreachCommandType }
func (CreateSlaBreachCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.SlaBreachSchemaName)
}

type CreateSlaBreachResult = dyn.OpResult[domain.SlaBreach]

var deleteSlaBreachCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "deleteSlaBreach"}

type DeleteSlaBreachCommand dyn.DeleteOneCommand

func (DeleteSlaBreachCommand) CqrsRequestType() cqrs.RequestType { return deleteSlaBreachCommandType }

type DeleteSlaBreachResult = dyn.OpResult[dyn.MutateResultData]

var getSlaBreachQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "getSlaBreach"}

type GetSlaBreachQuery dyn.GetOneQuery

func (GetSlaBreachQuery) CqrsRequestType() cqrs.RequestType { return getSlaBreachQueryType }

type GetSlaBreachResult = dyn.OpResult[domain.SlaBreach]

var slaBreachExistsQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "slaBreachExists"}

type SlaBreachExistsQuery dyn.ExistsQuery

func (SlaBreachExistsQuery) CqrsRequestType() cqrs.RequestType { return slaBreachExistsQueryType }

type SlaBreachExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchSlaBreachesQueryType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "searchSlaBreaches"}

type SearchSlaBreachesQuery dyn.SearchQuery

func (SearchSlaBreachesQuery) CqrsRequestType() cqrs.RequestType { return searchSlaBreachesQueryType }

type SearchSlaBreachesResultData = dyn.PagedResultData[domain.SlaBreach]
type SearchSlaBreachesResult = dyn.OpResult[SearchSlaBreachesResultData]

var updateSlaBreachCommandType = cqrs.RequestType{Module: "helpdesk", Submodule: "slabreach", Action: "updateSlaBreach"}

type UpdateSlaBreachCommand struct{ domain.SlaBreach }

func (UpdateSlaBreachCommand) CqrsRequestType() cqrs.RequestType { return updateSlaBreachCommandType }
func (UpdateSlaBreachCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.SlaBreachSchemaName)
}

type UpdateSlaBreachResult = dyn.OpResult[dyn.MutateResultData]
