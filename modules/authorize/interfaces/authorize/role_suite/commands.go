package role_suite

import (
	"time"


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
	req = (*CreateRoleSuiteCommand)(nil)
	req = (*UpdateRoleSuiteCommand)(nil)
	req = (*DeleteRoleSuiteCommand)(nil)
	req = (*UpdateRoleSuiteCommand)(nil)
	req = (*DeleteRoleSuiteCommand)(nil)
	req = (*GetRoleSuiteByIdQuery)(nil)
	req = (*GetRoleSuitesBySubjectQuery)(nil)
	req = (*ExistUserWithRoleSuiteQuery)(nil)
	req = (*SearchRoleSuitesCommand)(nil)
	req = (*AddRemoveUserCommand)(nil)
	util.Unused(req)
}

// START: CreateRoleSuiteCommand
var createRoleSuiteCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "create",
}

type CreateRoleSuiteCommand struct {
	Name                 string   `json:"name"`
	Description          *string  `json:"description,omitempty"`
	OwnerType            string   `json:"ownerType"`
	OwnerRef             model.Id `json:"ownerRef"`
	IsRequestable        *bool    `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool    `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool    `json:"isRequiredComment,omitempty"`
	CreatedBy            model.Id `json:"createdBy"`

	RoleIds []model.Id `json:"roleIds,omitempty"`
}

func (CreateRoleSuiteCommand) CqrsRequestType() cqrs.RequestType {
	return createRoleSuiteCommandType
}

type CreateRoleSuiteResult = crud.OpResult[*domain.RoleSuite]

// END: CreateRoleSuiteCommand

// START: UpdateRoleSuiteCommand
var updateRoleSuiteCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "update",
}

type UpdateRoleSuiteCommand struct {
	Id          model.Id   `param:"id" json:"id"`
	Etag        model.Etag `json:"etag"`
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`

	RoleIds []model.Id `json:"roleIds,omitempty"`
}

func (UpdateRoleSuiteCommand) CqrsRequestType() cqrs.RequestType {
	return updateRoleSuiteCommandType
}

type UpdateRoleSuiteResult = crud.OpResult[*domain.RoleSuite]

// END: UpdateRoleSuiteCommand

// START: DeleteRoleSuiteCommand
var deleteRoleSuiteCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "delete",
}

type DeleteRoleSuiteCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (DeleteRoleSuiteCommand) CqrsRequestType() cqrs.RequestType {
	return deleteRoleSuiteCommandType
}

func (this DeleteRoleSuiteCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteRoleSuiteResult = crud.DeletionResult

// END: DeleteRoleSuiteCommand

// START: GetRoleSuiteByIdQuery
var getRoleSuiteByIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "getRoleSuiteById",
}

type GetRoleSuiteByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetRoleSuiteByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getRoleSuiteByIdQueryType
}

type GetRoleSuiteByIdResult = crud.OpResult[*domain.RoleSuite]

// END: GetRoleSuiteByIdQuery

// START: GetRoleSuiteByNameCommand
var getRoleSuiteByNameCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "getRoleSuiteByName",
}

type GetRoleSuiteByNameCommand struct {
	Name string `param:"name" json:"name"`
}

func (GetRoleSuiteByNameCommand) CqrsRequestType() cqrs.RequestType {
	return getRoleSuiteByNameCommandType
}

type GetRoleSuiteByNameResult = crud.OpResult[*domain.RoleSuite]

// END: GetRoleSuiteByNameCommand

// START: SearchRoleSuitesCommand
var searchRoleSuitesCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "list",
}

type SearchRoleSuitesCommand struct {
	Page  *int    `json:"page" query:"page"`
	Size  *int    `json:"size" query:"size"`
	Graph *string `json:"graph" query:"graph"`
}

func (SearchRoleSuitesCommand) CqrsRequestType() cqrs.RequestType {
	return searchRoleSuitesCommandType
}

func (this *SearchRoleSuitesCommand) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchRoleSuitesCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchRoleSuitesResultData = crud.PagedResult[domain.RoleSuite]
type SearchRoleSuitesResult = crud.OpResult[*SearchRoleSuitesResultData]

// END: SearchRoleSuitesCommand

// START: GetRoleSuitesBySubjectQuery
var getRoleSuitesBySubjectQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "getRoleSuitesBySubject",
}

type GetRoleSuitesBySubjectQuery struct {
	SubjectType domain.EntitlementAssignmentSubjectType `param:"subjectType" json:"subjectType"`
	SubjectRef  string                                  `param:"subjectRef" json:"subjectRef"`
}

func (this GetRoleSuitesBySubjectQuery) Validate() fault.ValidationErrors {
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

func (GetRoleSuitesBySubjectQuery) CqrsRequestType() cqrs.RequestType {
	return getRoleSuitesBySubjectQueryType
}

type GetRoleSuitesBySubjectResult = crud.OpResult[[]domain.RoleSuite]

// END: GetRoleSuitesBySubjectQuery

// START: ExistUserWithRoleSuiteQuery
var existUserWithRoleSuiteQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "existUserWithRoleSuite",
}

type ExistUserWithRoleSuiteQuery struct {
	ReceiverType domain.ReceiverType `param:"receiverType" json:"receiverType"`
	ReceiverId   model.Id              `param:"receiverId" json:"receiverId"`
}

func (ExistUserWithRoleSuiteQuery) CqrsRequestType() cqrs.RequestType {
	return existUserWithRoleSuiteQueryType
}

type ExistUserWithRoleSuiteResult = crud.OpResult[bool]

// END: ExistUserWithRoleSuiteQuery

// START: AddRemoveUser
var addRemoveUserCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role_suite",
	Action:    "addRemoveUser",
}

type AddRemoveUserCommand struct {
	Id           model.Id            `param:"id" json:"id"`
	ApproverID   model.Id            `json:"approverId"`
	ReceiverID   model.Id            `json:"receiverId"`
	ReceiverType domain.ReceiverType `json:"receiverType"`
	Add          bool                `json:"add"`
}

func (AddRemoveUserCommand) CqrsRequestType() cqrs.RequestType {
	return addRemoveUserCommandType
}

func (this AddRemoveUserCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdValidateRule(&this.ApproverID, true),
		model.IdValidateRule(&this.ReceiverID, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type AddRemoveUserResultData struct {
	Id        model.Id   `json:"id"`
	Etag      model.Etag `json:"etag"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type AddRemoveUserResult = crud.OpResult[*AddRemoveUserResultData]

// END: AddRemoveUser
