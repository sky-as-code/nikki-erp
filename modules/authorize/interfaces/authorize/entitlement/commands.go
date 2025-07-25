package entitlement

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
	req = (*CreateEntitlementCommand)(nil)
	req = (*EntitlementExistsCommand)(nil)
	req = (*UpdateEntitlementCommand)(nil)
	req = (*DeleteHardEntitlementCommand)(nil)
	req = (*GetEntitlementByIdQuery)(nil)
	req = (*GetEntitlementByNameQuery)(nil)
	req = (*GetAllEntitlementByIdsQuery)(nil)
	req = (*SearchEntitlementsQuery)(nil)
	util.Unused(req)
}

// START: CreateEntitlementCommand
var createEntitlementCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "create",
}

type CreateEntitlementCommand struct {
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ActionId    *model.Id `json:"actionId,omitempty"`
	ResourceId  *model.Id `json:"resourceId,omitempty"`
	ScopeRef    *model.Id `json:"scopeRef,omitempty"`
	ActionExpr  string    `json:"actionExpr"`
	CreatedBy   string    `json:"createdBy"`
}

func (CreateEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return createEntitlementCommandType
}

type CreateEntitlementResult = crud.OpResult[*domain.Entitlement]

// END: CreateEntitlementCommand

// START: EntitlementExistsCommand
var existsCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "exists",
}

type EntitlementExistsCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (EntitlementExistsCommand) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

func (this EntitlementExistsCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type EntitlementExistsResult = crud.OpResult[bool]

// END: EntitlementExistsCommand

// START: UpdateEntitlementCommand
var updateEntitlementCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "update",
}

type UpdateEntitlementCommand struct {
	Id   model.Id   `param:"id" json:"id"`
	Etag model.Etag `json:"etag"`

	Description *string `json:"description,omitempty"`
}

func (UpdateEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return updateEntitlementCommandType
}

type UpdateEntitlementResult = crud.OpResult[*domain.Entitlement]

// END: UpdateEntitlementCommand

// START: DeleteHardEntitlementCommand
var deleteHardEntitlementCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "deleteHard",
}

type DeleteHardEntitlementCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (DeleteHardEntitlementCommand) CqrsRequestType() cqrs.RequestType {
	return deleteHardEntitlementCommandType
}

func (this DeleteHardEntitlementCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteHardEntitlementResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteHardEntitlementResult = crud.DeletionResult

// END: DeleteHardEntitlementCommand

// START: GetEntitlementByIdQuery
var getEntitlementByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "getById",
}

type GetEntitlementByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetEntitlementByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetEntitlementByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getEntitlementByIdQueryType
}

type GetEntitlementByIdResult = crud.OpResult[*domain.Entitlement]

// END: GetEntitlementByIdQuery

// START: GetEntitlementByNameQuery
var getEntitlementByNameQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "getByName",
}

type GetEntitlementByNameQuery struct {
	Name string `param:"name" json:"name"`
}

func (GetEntitlementByNameQuery) CqrsRequestType() cqrs.RequestType {
	return getEntitlementByNameQueryType
}

type GetEntitlementByNameResult = crud.OpResult[*domain.Entitlement]

// END: GetEntitlementByNameQuery

// START: GetAllEntitlementByIdsQuery
var getAllEntitlementByIdsQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "getAllByIds",
}

type GetAllEntitlementByIdsQuery struct {
	Ids []model.Id `param:"ids" json:"ids"`
}

func (GetAllEntitlementByIdsQuery) CqrsRequestType() cqrs.RequestType {
	return getAllEntitlementByIdsQueryType
}

func (this GetAllEntitlementByIdsQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRuleMulti(&this.Ids, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetAllEntitlementByIdsResult = crud.OpResult[[]*domain.Entitlement]

// END: GetAllEntitlementByIdsQuery

// START: SearchEntitlementsQuery
var searchEntitlementsQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "list",
}

type SearchEntitlementsQuery struct {
	Page  *int    `json:"page" query:"page"`
	Size  *int    `json:"size" query:"size"`
	Graph *string `json:"graph" query:"graph"`
}

func (SearchEntitlementsQuery) CqrsRequestType() cqrs.RequestType {
	return searchEntitlementsQueryType
}

func (this *SearchEntitlementsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchEntitlementsQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchEntitlementsResultData = crud.PagedResult[*domain.Entitlement]
type SearchEntitlementsResult = crud.OpResult[*SearchEntitlementsResultData]

// END: SearchEntitlementsQuery
