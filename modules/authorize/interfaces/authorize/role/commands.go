package role

import (
	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateRoleCommand)(nil)
	req = (*GetRoleByNameCommand)(nil)
	req = (*GetRoleByIdQuery)(nil)
	req = (*SearchRolesQuery)(nil)
	util.Unused(req)
}

// START: CreateRoleCommand
var createRoleCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "create",
}

type CreateRoleCommand struct {
	Name                 string  `json:"name"`
	Description          *string `json:"description,omitempty"`
	OwnerType            string  `json:"ownerType"`
	OwnerRef             string  `json:"ownerRef"`
	IsRequestable        bool    `json:"isRequestable"`
	IsRequiredAttachment bool    `json:"isRequiredAttachment"`
	IsRequiredComment    bool    `json:"isRequiredComment"`
	CreatedBy            string  `json:"createdBy"`

	Entitlements []*model.Id `json:"entitlements,omitempty"`
}

func (CreateRoleCommand) Type() cqrs.RequestType {
	return createRoleCommandType
}

type CreateRoleResult model.OpResult[*domain.Role]

// END: CreateRoleCommand

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

// START: GetRoleByIdQuery
var getRoleByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "getById",
}

type GetRoleByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetRoleByIdQuery) Type() cqrs.RequestType {
	return getRoleByIdQueryType
}

func (this GetRoleByIdQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetRoleByIdResult model.OpResult[*domain.Role]

// END: GetRoleByIdQuery

// START: GetRoleByNameCommand
var getRoleByNameCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "getByName",
}

type GetRoleByNameCommand struct {
	Name string `param:"name" json:"name"`
}

func (GetRoleByNameCommand) Type() cqrs.RequestType {
	return getRoleByNameCommandType
}

type GetRoleByNameResult model.OpResult[*domain.Role]

// END: GetRoleByNameCommand

// START: SearchRolesQuery
var searchRolesQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "list",
}

type SearchRolesQuery struct {
	Page             *int    `json:"page" query:"page"`
	Size             *int    `json:"size" query:"size"`
	Graph            *string `json:"graph" query:"graph"`
}

func (SearchRolesQuery) Type() cqrs.RequestType {
	return searchRolesQueryType
}

func (this *SearchRolesQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchRolesQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.PageIndexValidateRule(&this.Page),
		model.PageSizeValidateRule(&this.Size),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchRolesResultData = crud.PagedResult[*domain.Role]
type SearchRolesResult model.OpResult[*SearchRolesResultData]

// END: SearchRolesQuery
