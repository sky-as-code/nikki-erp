package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type UserDto struct {
	Id          string     `json:"id"`
	AvatarUrl   *string    `json:"avatarUrl,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	DisplayName string     `json:"displayName"`
	Email       string     `json:"email"`
	Etag        string     `json:"etag"`
	Status      string     `json:"status"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`

	Groups    []GroupDto          `json:"groups,omitempty"`
	Hierarchy []HierarchyLevelDto `json:"hierarchy,omitempty"`
	Orgs      []OrganizationDto   `json:"orgs,omitempty"`
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

	this.Hierarchy = array.Map(user.Hierarchy, func(hierarchy domain.HierarchyLevel) HierarchyLevelDto {
		hierarchyResp := HierarchyLevelDto{}
		hierarchyResp.FromHierarchyLevel(hierarchy)
		return hierarchyResp
	})

	this.Orgs = array.Map(user.Orgs, func(org domain.Organization) OrganizationDto {
		orgResp := OrganizationDto{}
		orgResp.FromOrg(org)
		return orgResp
	})
}

type CreateUserRequest = it.CreateUserCommand2
type CreateUserResponse = httpserver.RestCreateResponse

type UpdateUserRequest = it.UpdateUserCommand
type UpdateUserResponse = httpserver.RestUpdateResponse

type DeleteUserRequest = it.DeleteUserCommand
type DeleteUserResponse = httpserver.RestDeleteResponse

type GetUserByIdRequest = it.GetUser
type GetUserByIdResponse = UserDto

type GetUserContextRequest = it.GetUserContextQuery
type GetUserContextResponse = it.GetUserContextResult

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

type UserExistsMultiRequest = it.UserExistsMultiQuery
type UserExistsMultiResponse = it.ExistsMultiResultData

// UserEntityDto is the response DTO for dynamic-entity user operations.
// JSON field names use snake_case to match the dynamic entity schema column names.
type UserEntityDto struct {
	Id          string  `json:"id"`
	DisplayName *string `json:"display_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	AvatarUrl   *string `json:"avatar_url,omitempty"`
	Status      *string `json:"status,omitempty"`
	CreatedAt   *string `json:"created_at,omitempty"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
	Etag        *string `json:"etag,omitempty"`
}

type SearchUsers2Response struct {
	Items []schema.DynamicFields `json:"items"`
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

// func toSearchUsers2Response(items []domain.UserEntity) SearchUsers2Response {
// 	dtos := make([]UserEntityDto, len(items))
// 	for i, item := range items {
// 		dto, _ := modelmapper.MapToStruct[*UserEntityDto](item.GetFieldData())
// 		if dto != nil {
// 			dtos[i] = *dto
// 		}
// 	}
// 	return SearchUsers2Response{Items: dtos, Total: len(dtos)}
// }

type UpdateUser2Request = it.UpdateUserCommand2
type UpdateUser2Response = httpserver.RestUpdateResponse

type SearchUsers2Request = it.SearchUsersQuery2

type ArchiveUser2Request = it.ArchiveUserCommand2
type ArchiveUser2Response = httpserver.RestArchivedResponse
