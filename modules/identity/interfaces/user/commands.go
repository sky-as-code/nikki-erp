package user

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateUserCommand)(nil)
	req = (*UpdateUserCommand)(nil)
	req = (*DeleteUserCommand)(nil)
	req = (*GetUserByIdQuery)(nil)
	util.Unused(req)
}

var createUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	DisplayName        string    `json:"displayName"`
	Email              string    `json:"email"`
	IsActive           bool      `json:"isActive"`
	MustChangePassword bool      `json:"mustChangePassword"`
	Password           string    `json:"password"`
	OrgId              *model.Id `json:"orgId,omitempty"`
}

func (CreateUserCommand) Type() cqrs.RequestType {
	return createUserCommandType
}

type CreateUserResult model.OpResult[*domain.User]

var updateUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	Id                 model.Id   `param:"id" json:"id"`
	AvatarUrl          *string    `json:"avatarUrl,omitempty"`
	DisplayName        *string    `json:"displayName,omitempty"`
	Email              *string    `json:"email,omitempty"`
	Etag               model.Etag `json:"etag,omitempty"`
	IsActive           *bool      `json:"isActive,omitempty"`
	MustChangePassword *bool      `json:"mustChangePassword,omitempty"`
	Password           *string    `json:"password,omitempty"`
	OrgId              *model.Id  `json:"orgId,omitempty"`
}

func (UpdateUserCommand) Type() cqrs.RequestType {
	return updateUserCommandType
}

type UpdateUserResult model.OpResult[*domain.User]

var deleteUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "delete",
}

type DeleteUserCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeleteUserCommand) Type() cqrs.RequestType {
	return deleteUserCommandType
}

func (this DeleteUserCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteUserResultData struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteUserResult model.OpResult[*DeleteUserResultData]

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

type UserExistsCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (UserExistsCommand) Type() cqrs.RequestType {
	return existsCommandType
}

func (this UserExistsCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UserExistsResult model.OpResult[bool]

var existsMultiCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "existsMulti",
}

type UserExistsMultiCommand struct {
	Ids []model.Id `json:"ids"`
}

func (UserExistsMultiCommand) Type() cqrs.RequestType {
	return existsMultiCommandType
}

func (this UserExistsMultiCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRuleMulti(&this.Ids, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ExistsMultiResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"notExisting"`
}

type UserExistsMultiResult model.OpResult[*ExistsMultiResultData]

var getUserByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUserByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetUserByIdQuery) Type() cqrs.RequestType {
	return getUserByIdQueryType
}

func (this GetUserByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetUserByIdResult model.OpResult[*domain.User]

var searchUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "search",
}

type SearchUsersCommand struct {
	Page            *int    `json:"page" query:"page"`
	Size            *int    `json:"size" query:"size"`
	Graph           *string `json:"graph" query:"graph"`
	WithGroups      bool    `json:"withGroups" query:"withGroups"`
	WithOrgs        bool    `json:"withOrgs" query:"withOrgs"`
	WithHierarchies bool    `json:"withHierarchies" query:"withHierarchies"`
}

func (SearchUsersCommand) Type() cqrs.RequestType {
	return searchUsersCommandType
}

func (this *SearchUsersCommand) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchUsersCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.PageIndexValidateRule(&this.Page),
		model.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchUsersResultData = crud.PagedResult[domain.User]
type SearchUsersResult model.OpResult[*SearchUsersResultData]
