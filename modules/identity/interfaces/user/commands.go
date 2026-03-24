package user

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	dCrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateUserCommand)(nil)
	req = (*UpdateUserCommand)(nil)
	req = (*DeleteUserCommand)(nil)
	req = (*GetUser)(nil)
	req = (*GetUserByEmailQuery)(nil)
	req = (*SearchUsersQuery)(nil)
	req = (*UserExistsQuery)(nil)
	req = (*UserExistsMultiQuery)(nil)
	req = (*FindDirectApproverQuery)(nil)
	util.Unused(req)
}

var getUserContextQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserContext",
}

type GetUserContextQuery struct {
	UserId model.Id `json:"id" param:"id"`
}

func (GetUserContextQuery) CqrsRequestType() cqrs.RequestType {
	return getUserContextQueryType
}

func (this GetUserContextQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetUserContextResult struct {
	User        *domain.User                                       `json:"user,omitempty"`
	Hierachies  []domain.HierarchyLevel                            `json:"hierarchies,omitempty"`
	Orgs        []domain.Organization                              `json:"orgs,omitempty"`
	Permissions *map[string][]itAuthorize.ResourceScopePermissions `json:"permissions,omitempty"`
}

type GetUserContextResultData = corecrud.OpResult[*GetUserContextResult]

var createUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	DisplayName string    `json:"displayName" entity:"display_name"`
	Email       string    `json:"email"`
	OrgId       *model.Id `json:"orgId" entity:"org_id"`
}

func (CreateUserCommand) CqrsRequestType() cqrs.RequestType {
	return createUserCommandType
}

type CreateUserResult = corecrud.OpResult[*domain.User]

type CreateUserCommand2 struct {
	domain.UserEntity
}

func (CreateUserCommand2) CqrsRequestType() cqrs.RequestType {
	return createUserCommandType
}

func (this CreateUserCommand2) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserSchemaName)
}

type CreateUserResult2 = crud.OpResult[domain.UserEntity]

type UpdateUserCommand2 struct {
	domain.UserEntity
}

func (UpdateUserCommand2) CqrsRequestType() cqrs.RequestType {
	return updateUserCommandType
}

func (this UpdateUserCommand2) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserSchemaName)
}

type UpdateUserResult2 = crud.OpResult[domain.UserEntity]

type GetUserResult = crud.OpResult[domain.UserEntity]

var searchUsersQuery2Type = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "searchUsers2",
}

type SearchUsersQuery2 struct {
	Graph dmodel.SearchGraph `json:"graph"`
}

func (SearchUsersQuery2) CqrsRequestType() cqrs.RequestType {
	return searchUsersQuery2Type
}

type SearchUsersResult2 = crud.OpResult[crud.PagedResult[domain.UserEntity]]

var archiveUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "archive",
}

type ArchiveUserCommand2 struct {
	domain.UserEntity
}

func (ArchiveUserCommand2) CqrsRequestType() cqrs.RequestType {
	return archiveUserCommandType
}

func (this ArchiveUserCommand2) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetSchema(domain.UserSchemaName)
}

type ArchiveUserResult2 = crud.OpResult[domain.UserEntity]

var updateUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	Id          model.Id           `param:"id" json:"id"`
	AvatarUrl   *string            `json:"avatarUrl"`
	DisplayName *string            `json:"displayName"`
	Email       *string            `json:"email"`
	Etag        model.Etag         `json:"etag"`
	Status      *domain.UserStatus `json:"status"`
	ScopeRef    *model.Id          `query:"scopeRef" json:"scopeRef"`
}

func (UpdateUserCommand) CqrsRequestType() cqrs.RequestType {
	return updateUserCommandType
}

type UpdateUserResult = corecrud.OpResult[*domain.User]

var deleteUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "delete",
}

type DeleteUserCommand struct {
	Id       model.Id  `json:"id" param:"id"`
	ScopeRef *model.Id `query:"scopeRef" json:"scopeRef"`
}

func (DeleteUserCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUserCommandType
}

func (this DeleteUserCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteUserResult = corecrud.DeletionResult

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

type UserExistsQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (UserExistsQuery) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

func (this UserExistsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UserExistsResult = corecrud.OpResult[bool]

var existsMultiCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "existsMulti",
}

type UserExistsMultiQuery struct {
	Ids   []model.Id `json:"ids"`
	OrgId *model.Id  `json:"orgId"`
}

func (UserExistsMultiQuery) CqrsRequestType() cqrs.RequestType {
	return existsMultiCommandType
}

func (this UserExistsMultiQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRuleMulti(&this.Ids, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ExistsMultiResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"notExisting"`
}

type UserExistsMultiResult = corecrud.OpResult[*ExistsMultiResultData]

var getUserByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUser struct {
	dCrud.GetOneQueryBase

	Id            model.Id           `json:"id" param:"id"`
	Status        *domain.UserStatus `json:"status" query:"status"`
	WithGroup     bool               `json:"with_group" query:"with_group"`
	WithHierarchy bool               `json:"with_hierarchy" query:"with_hierarchy"`
	WithOrg       bool               `json:"with_org" query:"with_org"`
	ScopeRef      *model.Id          `json:"scope_ref" query:"scope_ref"`
}

func (GetUser) CqrsRequestType() cqrs.RequestType {
	return getUserByIdQueryType
}

func (this GetUser) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this GetUser) GetSchema() *dmodel.ModelSchema {
	fullSchema := dmodel.GetSchema(domain.UserSchemaName)
	return dmodel.GetOrRegisterSchema(
		dmodel.DefineEntity("identity.user.getUserById").
			Field(dmodel.CloneField(fullSchema, basemodel.FieldId)).
			Field(dmodel.CloneField(fullSchema, domain.UserFieldStatus).IsRequired(false).Default(nil)).
			Extend(basemodel.GetOneQuerySchemaBuilder()).
			Build(),
	)
}

func (this GetUser) GetFieldData() dmodel.DynamicFields {
	fields := dmodel.DynamicFields{
		basemodel.FieldId:      this.Id,
		domain.UserFieldStatus: this.Status,
	}
	return fields
}

type GetUserByIdResult = corecrud.OpResult[*domain.User]

var getUserByEmailQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserByEmail",
}

type GetUserByEmailQuery struct {
	Email  string             `param:"email" json:"email"`
	Status *domain.UserStatus `json:"status"`
}

func (GetUserByEmailQuery) CqrsRequestType() cqrs.RequestType {
	return getUserByEmailQueryType
}

func (this GetUserByEmailQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Email,
			val.NotEmpty,
			val.IsEmail,
			val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetUserByEmailResult = corecrud.OpResult[*domain.User]

var mustGetActiveUserQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "mustGetActiveUser",
}

type MustGetActiveUserQuery struct {
	Id    *string `json:"id"`
	Email *string `json:"email"`
}

func (MustGetActiveUserQuery) CqrsRequestType() cqrs.RequestType {
	return mustGetActiveUserQueryType
}

func (this MustGetActiveUserQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, this.Email == nil),
		val.Field(&this.Email,
			val.NotNilWhen(this.Id == nil),
			val.When(this.Email != nil,
				val.NotEmpty,
				val.IsEmail,
				val.Length(5, model.MODEL_RULE_USERNAME_LENGTH),
			),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type MustGetActiveUserResult = corecrud.OpResult[*domain.User]

var searchUsersQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "search",
}

type SearchUsersQuery struct {
	corecrud.SearchQuery

	WithGroups    bool      `json:"withGroups" query:"withGroups"`
	WithOrgs      bool      `json:"withOrgs" query:"withOrgs"`
	WithHierarchy bool      `json:"withHierarchy" query:"withHierarchy"`
	ScopeRef      *model.Id `json:"scopeRef" query:"scopeRef"`
}

func (SearchUsersQuery) CqrsRequestType() cqrs.RequestType {
	return searchUsersQueryType
}

func (this SearchUsersQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdPtrValidateRule(&this.ScopeRef, false))

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchUsersResultData = corecrud.PagedResult[domain.User]
type SearchUsersResult = corecrud.OpResult[*SearchUsersResultData]

var findDirectApproverQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "findDirectApprover",
}

type FindDirectApproverQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (FindDirectApproverQuery) CqrsRequestType() cqrs.RequestType {
	return findDirectApproverQueryType
}

func (this FindDirectApproverQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type FindDirectApproverResult = corecrud.OpResult[[]domain.User]
