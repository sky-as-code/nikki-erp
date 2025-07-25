package resource

import (
	"regexp"
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
	req = (*CreateResourceCommand)(nil)
	req = (*UpdateResourceCommand)(nil)
	req = (*DeleteHardResourceCommand)(nil)
	req = (*GetResourceByNameQuery)(nil)
	req = (*SearchResourcesQuery)(nil)
	util.Unused(req)
}

// START: CreateResourceCommand
var createResourceCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "create",
}

type CreateResourceCommand struct {
	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ResourceType string  `json:"resourceType"`
	ResourceRef  string  `json:"resourceRef"`
	ScopeType    string  `json:"scopeType"`
}

func (CreateResourceCommand) CqrsRequestType() cqrs.RequestType {
	return createResourceCommandType
}

type CreateResourceResult = crud.OpResult[*domain.Resource]

// END: CreateResourceCommand

// START: UpdateResourceCommand
var updateResourceCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "update",
}

type UpdateResourceCommand struct {
	Id   model.Id   `param:"id" json:"id"`
	Etag model.Etag `json:"etag"`

	Description *string `json:"description,omitempty"`
}

func (UpdateResourceCommand) CqrsRequestType() cqrs.RequestType {
	return updateResourceCommandType
}

type UpdateResourceResult = crud.OpResult[*domain.Resource]

// END: UpdateResourceCommand

// START: DeleteHardResourceCommand
var deleteHardResourceCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "delete",
}

type DeleteHardResourceCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeleteHardResourceCommand) CqrsRequestType() cqrs.RequestType {
	return deleteHardResourceCommandType
}

func (this DeleteHardResourceCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteHardResourceResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteHardResourceResult = crud.DeletionResult

// END: DeleteHardResourceCommand

// START: GetResourceByIdQuery
var getResourceByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "getResourceById",
}

type GetResourceByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetResourceByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getResourceByIdQueryType
}

// END: GetResourceByIdQuery

// START: GetResourceByNameQuery
var getResourceByNameQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "getResourceByName",
}

type GetResourceByNameQuery struct {
	Name string `param:"name" json:"name"`
}

func (GetResourceByNameQuery) CqrsRequestType() cqrs.RequestType {
	return getResourceByNameQueryType
}

func (this *GetResourceByNameQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.Name,
			validator.NotEmpty,
			validator.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
			validator.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
	}

	return validator.ApiBased.ValidateStruct(this, rules...)
}

type GetResourceByNameResult = crud.OpResult[*domain.Resource]

// END: GetResourceByNameQuery

// START: SearchResourcesQuery
var searchResourcesQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "resource",
	Action:    "list",
}

type SearchResourcesQuery struct {
	Page        *int    `json:"page" query:"page"`
	Size        *int    `json:"size" query:"size"`
	Graph       *string `json:"graph" query:"graph"`
	WithActions bool    `json:"withActions" query:"withActions"`
}

func (SearchResourcesQuery) CqrsRequestType() cqrs.RequestType {
	return searchResourcesQueryType
}

func (this *SearchResourcesQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchResourcesQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchResourcesResultData = crud.PagedResult[domain.Resource]
type SearchResourcesResult = crud.OpResult[*SearchResourcesResultData]

// END: SearchResourcesQuery
