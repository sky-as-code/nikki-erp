package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
	"github.com/thoas/go-funk"
)

type CreateUserRequest = it.CreateUserCommand
type CreateUserResponse = GetUserByIdResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = GetUserByIdResponse

type DeleteUserRequest = it.DeleteUserCommand

type DeleteUserResponse struct {
	DeletedAt int64 `json:"deletedAt,omitempty"`
}

type GetUserByIdRequest = it.GetUserByIdQuery

type GetUserByIdResponse struct {
	Id          model.Id   `json:"id,omitempty"`
	AvatarUrl   *string    `json:"avatarUrl,omitempty"`
	CreatedAt   int64      `json:"createdAt,omitempty"`
	DisplayName string     `json:"displayName,omitempty"`
	Email       string     `json:"email,omitempty"`
	Etag        model.Etag `json:"etag,omitempty"`
	Status      string     `json:"status,omitempty"`
	UpdatedAt   *int64     `json:"updatedAt,omitempty"`
}

func (this *GetUserByIdResponse) FromUser(user domain.User) {
	this.Id = *user.Id
	this.AvatarUrl = user.AvatarUrl
	this.CreatedAt = user.CreatedAt.UnixMilli()
	this.DisplayName = *user.DisplayName
	this.Email = *user.Email
	this.Etag = *user.Etag
	this.Status = user.Status.String()
	this.UpdatedAt = safe.GetTimeUnixMilli(user.UpdatedAt)
}

type SearchUsersRequest = it.SearchUsersCommand

type SearchUsersResponseItem struct {
	Id          model.Id                     `json:"id"`
	DisplayName string                       `json:"displayName"`
	Email       string                       `json:"email"`
	LockedUntil *int64                       `json:"lockedUntil,omitempty"`
	Status      domain.UserStatus            `json:"status"`
	Groups      []SearchUsersRespGroups      `json:"groups"`
	Hierarchies []SearchUsersRespHierarchies `json:"hierarchies"`
	Orgs        []GetGroupRespOrg            `json:"orgs"`
}

func (this *SearchUsersResponseItem) FromUser(user domain.User) {
	this.Id = *user.Id
	this.DisplayName = *user.DisplayName
	this.Email = *user.Email
	this.LockedUntil = safe.GetTimeUnix(user.LockedUntil)
	this.Status = *user.Status

	this.Groups = array.Map(user.Groups, func(group domain.Group) SearchUsersRespGroups {
		groupResp := SearchUsersRespGroups{}
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
}

type SearchUsersRespGroups struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *SearchUsersRespGroups) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
}

type SearchUsersRespHierarchies struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *SearchUsersRespHierarchies) FromHierarhy(hierarhy domain.HierarchyLevel) {
	this.Id = *hierarhy.Id
	this.Name = *hierarhy.Name
}

type SearchUsersRespOrgs struct {
	Id          model.Id   `json:"id"`
	DisplayName string     `json:"displayName"`
	Slug        model.Slug `json:"slug"`
}

func (this *SearchUsersRespOrgs) FromOrg(org domain.Organization) {
	this.Id = *org.Id
	this.DisplayName = *org.DisplayName
	this.Slug = *org.Slug
}

type SearchUsersResponse struct {
	Items []SearchUsersResponseItem `json:"items"`
	Total int                       `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}

func (this *SearchUsersResponse) FromResult(result *it.SearchUsersResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(user domain.User) SearchUsersResponseItem {
		item := SearchUsersResponseItem{}
		item.FromUser(user)
		return item
	}).([]SearchUsersResponseItem)
}

type UserExistsMultiRequest = it.UserExistsMultiCommand
type UserExistsMultiResponse = it.ExistsMultiResultData
