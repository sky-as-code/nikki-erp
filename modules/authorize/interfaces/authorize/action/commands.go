package action

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateActionCommand)(nil)
	req = (*UpdateActionCommand)(nil)
	req = (*DeleteHardActionCommand)(nil)
	req = (*GetActionByIdQuery)(nil)
	req = (*GetActionByNameCommand)(nil)
	req = (*SearchActionsCommand)(nil)
	util.Unused(req)
}

// START: CreateActionCommand
var createActionCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "create",
}

type CreateActionCommand struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	ResourceId  model.Id `json:"resourceId"`
	CreatedBy   string   `json:"createdBy"`
}

func (CreateActionCommand) CqrsRequestType() cqrs.RequestType {
	return createActionCommandType
}

type CreateActionResult = crud.OpResult[*domain.Action]

// END: CreateActionCommand

// START: UpdateActionCommand
var updateActionCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "update",
}

type UpdateActionCommand struct {
	Id   model.Id   `param:"id" json:"id"`
	Etag model.Etag `json:"etag"`

	Description *string `json:"description,omitempty"`
}

func (UpdateActionCommand) CqrsRequestType() cqrs.RequestType {
	return updateActionCommandType
}

type UpdateActionResult = crud.OpResult[*domain.Action]

// END: UpdateResourceCommand

// START: DeleteHardActionCommand
var deleteHardActionCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "deleteHard",
}

type DeleteHardActionCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (DeleteHardActionCommand) CqrsRequestType() cqrs.RequestType {
	return deleteHardActionCommandType
}

func (this DeleteHardActionCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteHardActionResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteHardActionResult = crud.DeletionResult

// END: DeleteHardActionCommand

// START: GetActionByIdQuery
var getActionByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "getActionById",
}

type GetActionByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetActionByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getActionByIdQueryType
}

func (this GetActionByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetActionByIdResult = crud.OpResult[*domain.Action]

// END: GetActionByIdQuery

// START: GetResourceByNameCommand
var getActionByNameCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "getActionByName",
}

type GetActionByNameCommand struct {
	Name       string   `param:"name" json:"name"`
	ResourceId model.Id `json:"resourceId"`
}

func (GetActionByNameCommand) CqrsRequestType() cqrs.RequestType {
	return getActionByNameCommandType
}

type GetActionByNameResult = crud.OpResult[*domain.Action]

// END: GetResourceByNameCommand

// START: SearchActionsCommand
var searchActionsCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "action",
	Action:    "list",
}

type SearchActionsCommand struct {
	Page  *int    `json:"page" query:"page"`
	Size  *int    `json:"size" query:"size"`
	Graph *string `json:"graph" query:"graph"`
}

func (SearchActionsCommand) CqrsRequestType() cqrs.RequestType {
	return searchActionsCommandType
}

func (this *SearchActionsCommand) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchActionsCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchActionsResultData = crud.PagedResult[domain.Action]
type SearchActionsResult = crud.OpResult[*SearchActionsResultData]

// END: SearchActionsCommand
