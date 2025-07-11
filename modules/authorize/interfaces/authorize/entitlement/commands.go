package entitlement

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateEntitlementCommand)(nil)
	// req = (*UpdateResourceCommand)(nil)
	req = (*GetEntitlementByNameCommand)(nil)
	// req = (*SearchResourcesCommand)(nil)
	util.Unused(req)
}

// START: CreateEntitlementCommand
var createEntitlementCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "create",
}

type CreateEntitlementCommand struct {
	ActionId    *model.Id                     `json:"actionId,omitempty"`
	ActionExpr  string                        `json:"actionExpr"`
	Name        string                        `json:"name"`
	Description *string                       `json:"description,omitempty"`
	ResourceId  *model.Id                     `json:"resourceId,omitempty"`
	// SubjectType domain.EntitlementSubjectType `json:"subjectType"`
	// SubjectRef  string                        `json:"subjectRef"`
	ScopeRef    *model.Id                     `json:"scopeRef,omitempty"`
	CreatedBy   string                        `json:"createdBy"`
}

func (CreateEntitlementCommand) Type() cqrs.RequestType {
	return createEntitlementCommandType
}

type CreateEntitlementResult model.OpResult[*domain.Entitlement]

// END: CreateEntitlementCommand

// // START: UpdateResourceCommand
// var updateResourceCommandType = cqrs.RequestType{
// 	Module:    "authorize",
// 	Submodule: "resource",
// 	Action:    "update",
// }

// type UpdateResourceCommand struct {
// 	Id          model.Id   `param:"id" json:"id"`
// 	Description *string    `json:"description,omitempty"`
// 	Etag        model.Etag `json:"etag,omitempty"`
// }

// func (UpdateResourceCommand) Type() cqrs.RequestType {
// 	return updateResourceCommandType
// }

// type UpdateResourceResult model.OpResult[*domain.Resource]

// // END: UpdateResourceCommand

// // START: GetResourceByIdQuery
// var getResourceByIdQueryType = cqrs.RequestType{
// 	Module:    "authorize",
// 	Submodule: "resource",
// 	Action:    "getResourceById",
// }

// type GetResourceByIdQuery struct {
// 	Id model.Id `param:"id" json:"id"`
// }

// func (GetResourceByIdQuery) Type() cqrs.RequestType {
// 	return getResourceByIdQueryType
// }

// // END: GetResourceByIdQuery

// START: GetEntitlementByNameCommand
var getEntitlementByNameCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "entitlement",
	Action:    "getByName",
}

type GetEntitlementByNameCommand struct {
	Name string `param:"name" json:"name"`
}

func (GetEntitlementByNameCommand) Type() cqrs.RequestType {
	return getEntitlementByNameCommandType
}

type GetEntitlementByNameResult model.OpResult[*domain.Entitlement]

// END: GetResourceByNameCommand

// // START: SearchResourcesCommand
// var searchResourcesCommandType = cqrs.RequestType{
// 	Module:    "authorize",
// 	Submodule: "resource",
// 	Action:    "list",
// }

// type SearchResourcesCommand struct {
// 	Page        *int    `json:"page" query:"page"`
// 	Size        *int    `json:"size" query:"size"`
// 	Graph       *string `json:"graph" query:"graph"`
// 	WithActions bool    `json:"withActions" query:"withActions"`
// }

// func (SearchResourcesCommand) Type() cqrs.RequestType {
// 	return searchResourcesCommandType
// }

// func (this *SearchResourcesCommand) SetDefaults() {
// 	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
// 	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
// }

// func (this SearchResourcesCommand) Validate() ft.ValidationErrors {
// 	rules := []*validator.FieldRules{
// 		model.PageIndexValidateRule(&this.Page),
// 		model.PageSizeValidateRule(&this.Size),
// 	}
// 	return validator.ApiBased.ValidateStruct(&this, rules...)
// }

// type SearchResourcesResultData = crud.PagedResult[domain.Resource]
// type SearchResourcesResult model.OpResult[*SearchResourcesResultData]

// // END: SearchResourcesCommand
