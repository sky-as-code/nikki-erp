package role

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateRoleCommand)(nil)
	req = (*UpdateRoleCommand)(nil)
	req = (*DeleteRoleHardCommand)(nil)
	req = (*GetRoleByNameCommand)(nil)
	req = (*GetRoleByIdQuery)(nil)
	req = (*SearchRolesQuery)(nil)
	req = (*GetRolesBySubjectQuery)(nil)
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

	EntitlementIds []model.Id `json:"entitlementIds,omitempty"`
}

func (CreateRoleCommand) CqrsRequestType() cqrs.RequestType {
	return createRoleCommandType
}

type CreateRoleResult = crud.OpResult[*domain.Role]

// END: CreateRoleCommand

// START: UpdateRoleCommand
var updateRoleCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "update",
}

type UpdateRoleCommand struct {
	Id          model.Id   `param:"id" json:"id"`
	Etag        model.Etag `json:"etag"`
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`

	EntitlementIds []model.Id `json:"entitlementIds,omitempty"`
}

func (UpdateRoleCommand) CqrsRequestType() cqrs.RequestType {
	return updateRoleCommandType
}

type UpdateRoleResult = crud.OpResult[*domain.Role]

// END: UpdateRoleCommand

// START: DeleteRoleHardCommand
var deleteRoleHardCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "deleteHard",
}

type DeleteRoleHardCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (DeleteRoleHardCommand) CqrsRequestType() cqrs.RequestType {
	return deleteRoleHardCommandType
}

func (this DeleteRoleHardCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteRoleHardResult = crud.DeletionResult

// END: DeleteRoleHardCommand

// START: GetRoleByIdQuery
var getRoleByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "getById",
}

type GetRoleByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetRoleByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getRoleByIdQueryType
}

func (this GetRoleByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetRoleByIdResult = crud.OpResult[*domain.Role]

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

func (GetRoleByNameCommand) CqrsRequestType() cqrs.RequestType {
	return getRoleByNameCommandType
}

type GetRoleByNameResult = crud.OpResult[*domain.Role]

// END: GetRoleByNameCommand

// START: SearchRolesQuery
var searchRolesQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "list",
}

type SearchRolesQuery struct {
	Page  *int    `json:"page" query:"page"`
	Size  *int    `json:"size" query:"size"`
	Graph *string `json:"graph" query:"graph"`
}

func (SearchRolesQuery) CqrsRequestType() cqrs.RequestType {
	return searchRolesQueryType
}

func (this *SearchRolesQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchRolesQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchRolesResultData = crud.PagedResult[domain.Role]
type SearchRolesResult = crud.OpResult[*SearchRolesResultData]

// END: SearchRolesQuery

// START: GetRolesBySubjectQuery
var getRolesBySubjectQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "getBySubject",
}

type GetRolesBySubjectQuery struct {
	SubjectType domain.EntitlementAssignmentSubjectType `param:"subjectType" json:"subjectType"`
	SubjectRef  string                                  `param:"subjectRef" json:"subjectRef"`
}

func (this GetRolesBySubjectQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.SubjectType,
			validator.NotEmpty,
			validator.OneOf(
				domain.EntitlementAssignmentSubjectTypeNikkiUser,
				domain.EntitlementAssignmentSubjectTypeNikkiGroup,
				domain.EntitlementAssignmentSubjectTypeNikkiRole,
				domain.EntitlementAssignmentSubjectTypeCustom,
			),
		),
		model.IdValidateRule(&this.SubjectRef, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (GetRolesBySubjectQuery) CqrsRequestType() cqrs.RequestType {
	return getRolesBySubjectQueryType
}

type GetRolesBySubjectResult = crud.OpResult[[]domain.Role]

// END: GetRolesBySubjectQuery
