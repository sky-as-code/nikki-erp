package client

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

var userExistsRequestType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "exists",
}

// UserExistsQuery mirrors identity user.exists CQRS contract (no import of identity module).
type UserExistsQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (UserExistsQuery) CqrsRequestType() cqrs.RequestType {
	return userExistsRequestType
}

func (this UserExistsQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type UserExistsResult = crud.OpResult[bool]

var searchUsersRequestType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "search",
}

// UserEntity is a drive-local representation of identity domain.User.
// It's safe to keep it here because it only relies on JSON field names.
type UserEntity struct {
	Id          *model.Id   `json:"id,omitempty"`
	Etag        *model.Etag `json:"etag,omitempty"`
	AvatarUrl   *string     `json:"avatarUrl,omitempty"`
	DisplayName *string     `json:"displayName,omitempty"`
	Email       *string     `json:"email,omitempty"`

	HierarchyId *model.Id `json:"hierarchyId,omitempty"`
	OrgId       *model.Id `json:"orgId,omitempty"`
	Status      *string   `json:"status,omitempty"`
	ScopeRef    *model.Id `json:"scopeRef,omitempty"`

	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// SearchUsersQuery mirrors identity user.search CQRS contract (no import of identity module).
type SearchUsersQuery struct {
	crud.SearchQuery

	WithGroups    bool     `json:"withGroups" query:"withGroups"`
	WithOrgs      bool     `json:"withOrgs" query:"withOrgs"`
	WithHierarchy bool     `json:"withHierarchy" query:"withHierarchy"`
	ScopeRef      *model.Id `json:"scopeRef" query:"scopeRef"`
}

func (SearchUsersQuery) CqrsRequestType() cqrs.RequestType {
	return searchUsersRequestType
}

func (this SearchUsersQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdPtrValidateRule(&this.ScopeRef, false))

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type SearchUsersResultData = crud.PagedResult[UserEntity]
type SearchUsersResult = crud.OpResult[*SearchUsersResultData]

var getUserByIdRequestType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

// GetUserByIdQuery mirrors identity user.getUserById CQRS contract (no import of identity module).
type GetUserByIdQuery struct {
	Id model.Id `param:"id" json:"id"`

	// Use *string to avoid importing identity's domain.UserStatus.
	Status *string `json:"status" query:"status"`

	WithGroup     bool     `json:"withGroup" query:"withGroup"`
	WithHierarchy bool     `json:"withHierarchy" query:"withHierarchy"`
	WithOrg       bool     `json:"withOrg" query:"withOrg"`
	ScopeRef      *model.Id `json:"scopeRef" query:"scopeRef"`
}

func (GetUserByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getUserByIdRequestType
}

func (this GetUserByIdQuery) Validate() ft.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}

	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetUserByIdResult = crud.OpResult[*UserEntity]
