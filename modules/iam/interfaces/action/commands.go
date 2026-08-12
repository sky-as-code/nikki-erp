package action

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func init() {
	var req cqrs.Request
	req = (*CreateActionCommand)(nil)
	req = (*DeleteActionCommand)(nil)
	req = (*GetActionQuery)(nil)
	req = (*ActionExistsQuery)(nil)
	req = (*SearchActionsQuery)(nil)
	req = (*UpdateActionCommand)(nil)
	util.Unused(req)
}

var createActionCommandType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "createAction"}

type CreateActionCommand struct {
	domain.Action
}

func (CreateActionCommand) CqrsRequestType() cqrs.RequestType { return createActionCommandType }

func (CreateActionCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ActionSchemaName)
}

type CreateActionResult = dyn.OpResult[domain.Action]

var deleteActionCommandType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "deleteAction"}

type DeleteActionCommand struct {
	ActionId   string   `json:"action_id" param:"action_id"`
	ResourceId model.Id `json:"resource_id" param:"resource_id"`
}

func (DeleteActionCommand) CqrsRequestType() cqrs.RequestType { return deleteActionCommandType }

func (DeleteActionCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.delete_action_command",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("action_code").
					RequiredAlways()).
				Field(basemodel.DefineFieldId("resource_id").
					RequiredAlways())
		},
	)
}

type DeleteActionResult = dyn.OpResult[dyn.MutateResultData]

var getActionQueryType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "getAction"}

type GetActionQuery struct {
	ActionId   string   `json:"action_id" param:"action_id"`
	ResourceId model.Id `json:"resource_id" param:"resource_id"`
	Columns    []string `json:"columns" query:"columns"`
}

func (GetActionQuery) CqrsRequestType() cqrs.RequestType { return getActionQueryType }

func (GetActionQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.get_action_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("action_id").
					RequiredAlways()).
				Field(basemodel.DefineFieldId("resource_id").
					RequiredAlways()).
				Field(dyn.DefineFieldSearchFields())
		},
	)
}

type GetActionResult = dyn.OpResult[dyn.SingleResultData[domain.Action]]

var actionExistsQueryType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "actionExists"}

type ActionExistsQuery struct {
	ActionIds  []model.Id `json:"action_ids"`
	ResourceId model.Id   `json:"resource_id" param:"resource_id"`
}

func (ActionExistsQuery) CqrsRequestType() cqrs.RequestType { return actionExistsQueryType }

func (ActionExistsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.action_exists_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldIdArr("action_ids").
					RequiredAlways()).
				Field(basemodel.DefineFieldId("resource_id").
					RequiredAlways())
		},
	)
}

type ActionExistsResult = dyn.OpResult[dyn.ExistsResultData]

var searchActionsQueryType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "searchActions"}

type SearchActionsQuery struct {
	Columns         []string            `json:"columns" query:"columns"`
	Graph           *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page            int                 `json:"page" query:"page"`
	Size            int                 `json:"size" query:"size"`
	IncludeArchived *bool               `json:"include_archived" query:"include_archived"`
	ResourceId      model.Id            `json:"resource_id" param:"resource_id"`
}

func (SearchActionsQuery) CqrsRequestType() cqrs.RequestType { return searchActionsQueryType }

func (SearchActionsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.search_actions_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(dyn.DefineFieldSearchFields()).
				Field(dyn.DefineFieldSearchGraph()).
				Field(dyn.DefineFieldSearchPage()).
				Field(dyn.DefineFieldSearchSize()).
				Field(dyn.DefineFieldIncludeArchived()).
				Field(basemodel.DefineFieldId("resource_id").
					RequiredAlways())
		},
	)
}

type SearchActionsResultData = dyn.PagedResultData[domain.Action]
type SearchActionsResult = dyn.OpResult[SearchActionsResultData]

var updateActionCommandType = cqrs.RequestType{Module: "iam", Submodule: "action", Action: "updateAction"}

type UpdateActionCommand struct {
	domain.Action
}

func (UpdateActionCommand) CqrsRequestType() cqrs.RequestType { return updateActionCommandType }

func (UpdateActionCommand) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.ActionSchemaName)
}

type UpdateActionResult = dyn.OpResult[dyn.MutateResultData]
