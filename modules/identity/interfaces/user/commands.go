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
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateUserCommand)(nil)
	req = (*UpdateUserCommand)(nil)
	req = (*DeleteUserCommand)(nil)
	req = (*GetUserByIdQuery)(nil)
	req = (*ListUserStatusesQuery)(nil)
	req = (*SearchUsersQuery)(nil)
	req = (*UserExistsCommand)(nil)
	req = (*UserExistsMultiCommand)(nil)
	util.Unused(req)
}

var createUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	DisplayName        string     `json:"displayName"`
	Email              string     `json:"email"`
	MustChangePassword bool       `json:"mustChangePassword"`
	Password           string     `json:"password"`
	OrgIds             []model.Id `json:"orgIds"`
}

func (CreateUserCommand) CqrsRequestType() cqrs.RequestType {
	return createUserCommandType
}

type CreateUserResult = crud.OpResult[*domain.User]

var updateUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	Id                 model.Id   `param:"id" json:"id"`
	AvatarUrl          *string    `json:"avatarUrl"`
	DisplayName        *string    `json:"displayName"`
	Email              *string    `json:"email"`
	Etag               model.Etag `json:"etag"`
	MustChangePassword *bool      `json:"mustChangePassword"`
	Password           *string    `json:"password"`
	StatusId           *model.Id  `json:"statusId"`
	StatusValue        *string    `json:"statusValue"`
}

func (UpdateUserCommand) CqrsRequestType() cqrs.RequestType {
	return updateUserCommandType
}

type UpdateUserResult = crud.OpResult[*domain.User]

var deleteUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "delete",
}

type DeleteUserCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (DeleteUserCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUserCommandType
}

func (this DeleteUserCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteUserResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteUserResult = crud.DeletionResult

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

type UserExistsCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (UserExistsCommand) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

func (this UserExistsCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UserExistsResult = crud.OpResult[bool]

var existsMultiCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "existsMulti",
}

type UserExistsMultiCommand struct {
	Ids []model.Id `json:"ids"`
}

func (UserExistsMultiCommand) CqrsRequestType() cqrs.RequestType {
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

type UserExistsMultiResult = crud.OpResult[*ExistsMultiResultData]

var getUserByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUserByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (GetUserByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getUserByIdQueryType
}

func (this GetUserByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetUserByIdResult = crud.OpResult[*domain.User]

var searchUsersQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "search",
}

type SearchUsersQuery struct {
	Page            *int    `json:"page" query:"page"`
	Size            *int    `json:"size" query:"size"`
	Graph           *string `json:"graph" query:"graph"`
	WithGroups      bool    `json:"withGroups" query:"withGroups"`
	WithOrgs        bool    `json:"withOrgs" query:"withOrgs"`
	WithHierarchies bool    `json:"withHierarchies" query:"withHierarchies"`
}

func (SearchUsersQuery) CqrsRequestType() cqrs.RequestType {
	return searchUsersQueryType
}

func (this *SearchUsersQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchUsersQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchUsersResultData = crud.PagedResult[domain.User]
type SearchUsersResult = crud.OpResult[*SearchUsersResultData]

var listUserStatusesCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "listUserStatuses",
}

type ListUserStatusesQuery struct {
	enum.ListDerivedEnumsQuery
}

func (ListUserStatusesQuery) CqrsRequestType() cqrs.RequestType {
	return listUserStatusesCommandType
}

type ListIdentStatusesResultData = crud.PagedResult[domain.IdentityStatus]
type ListIdentStatusesResult = crud.OpResult[*ListIdentStatusesResultData]
