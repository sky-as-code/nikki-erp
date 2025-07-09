package role_suite

// import (
// 	"github.com/sky-as-code/nikki-erp/common/crud"
// 	ft "github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/safe"
// 	"github.com/sky-as-code/nikki-erp/common/util"
// 	"github.com/sky-as-code/nikki-erp/common/validator"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
// )

// func init() {
// 	// Assert interface implementation
// 	var req cqrs.Request
// 	req = (*CreateResourceCommand)(nil)
// 	req = (*UpdateResourceCommand)(nil)
// 	req = (*GetResourceByNameCommand)(nil)
// 	req = (*SearchResourcesCommand)(nil)
// 	util.Unused(req)
// }

// // START: CreateResourceCommand
// var createResourceCommandType = cqrs.RequestType{
// 	Module:    "authorize",
// 	Submodule: "resource",
// 	Action:    "create",
// }

// type CreateResourceCommand struct {
// 	Name         string `json:"name"`
// 	Description  string `json:"description"`
// 	ResourceType string `json:"resourceType"`
// 	ResourceRef  string `json:"resourceRef"`
// 	ScopeType    string `json:"scopeType"`
// }

// func (CreateResourceCommand) Type() cqrs.RequestType {
// 	return createResourceCommandType
// }

// type CreateResourceResult model.OpResult[*domain.Resource]

// // END: CreateResourceCommand

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

// // START: GetResourceByNameCommand
// var getResourceByNameCommandType = cqrs.RequestType{
// 	Module:    "authorize",
// 	Submodule: "resource",
// 	Action:    "getResourceByName",
// }

// type GetResourceByNameCommand struct {
// 	Name string `param:"name" json:"name"`
// }

// func (GetResourceByNameCommand) Type() cqrs.RequestType {
// 	return getResourceByNameCommandType
// }

// type GetResourceByNameResult model.OpResult[*domain.Resource]

// // END: GetResourceByNameCommand

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
