package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type CreateGroupRequest = it.CreateGroupCommand
type CreateGroupResponse = GetGroupByIdResponse

type DeleteGroupRequest = it.DeleteGroupCommand

type DeleteGroupResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

type UpdateGroupRequest = it.UpdateGroupCommand
type UpdateGroupResponse = GetGroupByIdResponse

type GetGroupByIdRequest = it.GetGroupByIdQuery

type GetGroupByIdResponse struct {
	Id          model.Id         `json:"id"`
	CreatedAt   int64            `json:"createdAt,omitempty"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Etag        model.Etag       `json:"etag"`
	Org         *GetGroupRespOrg `json:"org,omitempty"`
	UpdatedAt   *int64           `json:"updatedAt,omitempty"`
}

func (this *GetGroupByIdResponse) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.CreatedAt = group.CreatedAt.UnixMilli()
	this.Name = *group.Name
	this.Description = group.Description
	this.Etag = *group.Etag
	this.UpdatedAt = safe.GetTimeUnixMilli(group.UpdatedAt)
	if group.Org != nil {
		this.Org = &GetGroupRespOrg{}
		this.Org.FromOrg(group.Org)
	}
}

type GetGroupRespOrg struct {
	Id          model.Id   `json:"id"`
	DisplayName string     `json:"displayName"`
	Slug        model.Slug `json:"slug"`
}

func (this *GetGroupRespOrg) FromOrg(org *domain.Organization) {
	if org == nil {
		return
	}
	this.Id = *org.Id
	this.DisplayName = *org.DisplayName
	this.Slug = *org.Slug
}

type SearchGroupsRequest = it.SearchGroupsQuery

type SearchGroupsResponseItem struct {
	Id          model.Id         `json:"id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Org         *GetGroupRespOrg `json:"org,omitempty"`
}

func (this *SearchGroupsResponseItem) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
	this.Description = group.Description
	if group.Org != nil {
		this.Org = &GetGroupRespOrg{}
		this.Org.FromOrg(group.Org)
	}
}

type SearchGroupsResponse struct {
	Items []SearchGroupsResponseItem `json:"items"`
	Total int                        `json:"total"`
	Page  int                        `json:"page"`
	Size  int                        `json:"size"`
}

func (this *SearchGroupsResponse) FromResult(result *it.SearchGroupsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(group domain.Group) SearchGroupsResponseItem {
		item := SearchGroupsResponseItem{}
		item.FromGroup(group)
		return item
	})
}

type ManageUsersRequest = it.AddRemoveUsersCommand
type ManageUsersResponse struct {
	UpdatedAt int64 `json:"updatedAt"`
}

func (this *ManageUsersResponse) FromResult(result *it.AddRemoveUsersResultData) {
	this.UpdatedAt = result.UpdatedAt.UnixMilli()
}
