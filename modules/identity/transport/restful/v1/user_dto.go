package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type UserDto struct {
	Id          string  `json:"id"`
	AvatarUrl   *string `json:"avatarUrl,omitempty"`
	CreatedAt   int64   `json:"createdAt"`
	DisplayName string  `json:"displayName"`
	Email       string  `json:"email"`
	Etag        string  `json:"etag"`
	Status      string  `json:"status"`
	UpdatedAt   *int64  `json:"updatedAt,omitempty"`

	Groups      []GroupDto                   `json:"groups,omitempty"`
	Hierarchies []SearchUsersRespHierarchies `json:"hierarchies,omitempty"`
	Orgs        []GetGroupRespOrg            `json:"orgs,omitempty"`
}

func (this *UserDto) FromUser(user domain.User) {
	model.MustCopy(user.AuditableBase, this)
	model.MustCopy(user.ModelBase, this)
	model.MustCopy(user, this)

	this.Groups = array.Map(user.Groups, func(group domain.Group) GroupDto {
		groupResp := GroupDto{}
		groupResp.FromGroup(group)
		return groupResp
	})

	this.Hierarchies = array.Map(user.Hierarchies, func(hierarhy domain.HierarchyLevel) SearchUsersRespHierarchies {
		hierarhyResp := SearchUsersRespHierarchies{}
		hierarhyResp.FromHierarhy(hierarhy)
		return hierarhyResp
	})

	this.Orgs = array.Map(user.Orgs, func(org domain.Organization) GetGroupRespOrg {
		orgResp := GetGroupRespOrg{}
		orgResp.FromOrg(&org)
		return orgResp
	})
	if user.Status != nil {
		this.Status = *user.Status.Value
	}
}

type CreateUserRequest = it.CreateUserCommand
type CreateUserResponse = httpserver.RestCreateResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = httpserver.RestUpdateResponse

type DeleteUserRequest = it.DeleteUserCommand
type DeleteUserResponse = httpserver.RestDeleteResponse

type GetUserByIdRequest = it.GetUserByIdQuery
type GetUserByIdResponse = UserDto

type SearchUsersRequest = it.SearchUsersQuery

type SearchUsersRespGroups struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (this *SearchUsersRespGroups) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
}

// TODO: Replace with HierarchyDto
type SearchUsersRespHierarchies struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (this *SearchUsersRespHierarchies) FromHierarhy(hierarhy domain.HierarchyLevel) {
	this.Id = *hierarhy.Id
	this.Name = *hierarhy.Name
}

// TODO: Replace with OrganizationDto

type SearchUsersRespOrgs struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

func (this *SearchUsersRespOrgs) FromOrg(org domain.Organization) {
	this.Id = *org.Id
	this.DisplayName = *org.DisplayName
	this.Slug = *org.Slug
}

type SearchUsersResponse struct {
	Items []UserDto `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
}

func (this *SearchUsersResponse) FromResult(result *it.SearchUsersResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(user domain.User) UserDto {
		item := UserDto{}
		item.FromUser(user)
		return item
	})
}

type UserExistsMultiRequest = it.UserExistsMultiCommand
type UserExistsMultiResponse = it.ExistsMultiResultData

type ListUserStatusesRequest = it.ListUserStatusesQuery
type ListUserStatusesResponse = it.ListUserStatusesResultData
