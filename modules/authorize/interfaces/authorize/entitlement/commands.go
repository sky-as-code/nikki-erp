package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateEntitlementCommand)(nil)
	req = (*EntitlementExistsCommand)(nil)
	req = (*UpdateEntitlementCommand)(nil)
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

func (CreateEntitlementCommand) Type() cqrs.RequestType {
	return createEntitlementCommandType
}

type CreateEntitlementResult model.OpResult[*domain.Entitlement]

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

func (EntitlementExistsCommand) Type() cqrs.RequestType {
	return existsCommandType
}

func (this EntitlementExistsCommand) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type EntitlementExistsResult model.OpResult[bool]

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

func (UpdateEntitlementCommand) Type() cqrs.RequestType {
	return updateEntitlementCommandType
}

type UpdateEntitlementResult model.OpResult[*domain.Entitlement]

// END: UpdateEntitlementCommand

// START: GetEntitlementByIdQuery
var getEntitlementByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "getById",
}

type GetEntitlementByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetEntitlementByIdQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetEntitlementByIdQuery) Type() cqrs.RequestType {
	return getEntitlementByIdQueryType
}

type GetEntitlementByIdResult model.OpResult[*domain.Entitlement]

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

func (GetEntitlementByNameQuery) Type() cqrs.RequestType {
	return getEntitlementByNameQueryType
}

type GetEntitlementByNameResult model.OpResult[*domain.Entitlement]

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

func (GetAllEntitlementByIdsQuery) Type() cqrs.RequestType {
	return getAllEntitlementByIdsQueryType
}

type GetAllEntitlementByIdsResult model.OpResult[[]*domain.Entitlement]

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

func (SearchEntitlementsQuery) Type() cqrs.RequestType {
	return searchEntitlementsQueryType
}

func (this *SearchEntitlementsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchEntitlementsQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.PageIndexValidateRule(&this.Page),
		model.PageSizeValidateRule(&this.Size),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchEntitlementsResultData = crud.PagedResult[*domain.Entitlement]
type SearchEntitlementsResult model.OpResult[*SearchEntitlementsResultData]

// END: SearchEntitlementsQuery
