package role

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
	req = (*CreateRoleCommand)(nil)
	req = (*DeleteRoleHardCommand)(nil)
	req = (*UpdateRoleCommand)(nil)
	req = (*GetRoleByNameCommand)(nil)
	req = (*GetRoleByIdQuery)(nil)
	req = (*GetRolesBySubjectQuery)(nil)
	req = (*ExistUserWithRoleQuery)(nil)
	req = (*SearchRolesQuery)(nil)
	req = (*AddRemoveUserCommand)(nil)
	util.Unused(req)
}

// START: CreateRoleCommand
var createRoleCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "create",
}

type CreateRoleCommand struct {
	Name                 string    `json:"name"`
	Description          *string   `json:"description,omitempty"`
	OwnerType            string    `json:"ownerType"`
	OwnerRef             string    `json:"ownerRef"`
	IsRequestable        bool      `json:"isRequestable"`
	IsRequiredAttachment bool      `json:"isRequiredAttachment"`
	IsRequiredComment    bool      `json:"isRequiredComment"`
	CreatedBy            string    `json:"createdBy"`
	OrgId                *model.Id `json:"orgId,omitempty"`
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
	Name  string    `param:"name" json:"name"`
	OrgId *model.Id `json:"orgId,omitempty"`
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
	crud.SearchQuery
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

// START: ExistUserWithRoleQuery
var existUserWithRoleQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "existUserWithRole",
}

type ExistUserWithRoleQuery struct {
	ReceiverType domain.ReceiverType `param:"receiverType" json:"receiverType"`
	ReceiverId   model.Id            `param:"receiverId" json:"receiverId"`
	TargetId     model.Id            `param:"targetId" json:"targetId"`
}

func (ExistUserWithRoleQuery) CqrsRequestType() cqrs.RequestType {
	return existUserWithRoleQueryType
}

type ExistUserWithRoleResult = crud.OpResult[bool]

// END: ExistUserWithRoleQuery

// START: AddRemoveUser
var addRemoveUserCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "addRemoveUser",
}

type AddRemoveUserCommand struct {
	Id           model.Id            `param:"id" json:"id"`
	ApproverID   model.Id            `json:"approverId"`
	ReceiverID   model.Id            `json:"receiverId"`
	ReceiverType domain.ReceiverType `json:"receiverType"`
	Add          bool                `json:"add"`
}

func (this *AddRemoveUserCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdValidateRule(&this.ApproverID, true),
		model.IdValidateRule(&this.ReceiverID, true),
	}

	return validator.ApiBased.ValidateStruct(this, rules...)
}

func (AddRemoveUserCommand) CqrsRequestType() cqrs.RequestType {
	return addRemoveUserCommandType
}

type AddRemoveUserResultData struct {
	Id        model.Id   `json:"id"`
	Etag      model.Etag `json:"etag"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type AddRemoveUserResult = crud.OpResult[*AddRemoveUserResultData]

// END: AddRemoveUser

// START: AddEntitlementsCommand
var addEntitlementsCommand = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "addEntitlements",
}

type AddEntitlementsCommand struct {
	Id                model.Id           `param:"id" json:"id"`
	Etag              model.Etag         `json:"etag"`
	EntitlementInputs []EntitlementInput `json:"entitlementInputs"`
}

type EntitlementInput struct {
	EntitlementId model.Id  `json:"entitlementId"`
	ScopeRef      *model.Id `json:"scopeRef,omitempty"`
}

func (AddEntitlementsCommand) CqrsRequestType() cqrs.RequestType {
	return addEntitlementsCommand
}

func (this AddEntitlementsCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type AddEntitlementsResult = crud.OpResult[*domain.Role]

// END: AddEntitlementsCommand

// START: RemoveEntitlementsCommand
var removeEntitlementsCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "role",
	Action:    "removeEntitlements",
}

type RemoveEntitlementsCommand struct {
	Id                model.Id           `param:"id" json:"id"`
	Etag              model.Etag         `json:"etag"`
	EntitlementInputs []EntitlementInput `json:"entitlementInputs"`
}

func (RemoveEntitlementsCommand) CqrsRequestType() cqrs.RequestType {
	return removeEntitlementsCommandType
}

func (this RemoveEntitlementsCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type RemoveEntitlementsResult = crud.OpResult[*domain.Role]

// END: RemoveEntitlementsCommand
