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

	Groups []GroupDto `json:"groups,omitempty"`
	// Hierarchies []HierarchyDto    `json:"hierarchies,omitempty"`
	Orgs []OrganizationDto `json:"orgs,omitempty"`
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

	// this.Hierarchies = array.Map(user.Hierarchies, func(hierarhy domain.HierarchyLevel) SearchUsersRespHierarchies {
	// 	hierarhyResp := SearchUsersRespHierarchies{}
	// 	hierarhyResp.FromHierarhy(hierarhy)
	// 	return hierarhyResp
	// })

	this.Orgs = array.Map(user.Orgs, func(org domain.Organization) OrganizationDto {
		orgResp := OrganizationDto{}
		orgResp.FromOrg(org)
		return orgResp
	})
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

type SearchUsersResponse httpserver.RestSearchResponse[UserDto]

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
type ListUserStatusesResponse = it.ListIdentStatusesResultData
