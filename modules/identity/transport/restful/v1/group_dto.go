package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type GroupDto struct {
	Id          string           `json:"id"`
	CreatedAt   int64            `json:"createdAt"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Etag        string           `json:"etag"`
	Org         *GetGroupRespOrg `json:"org,omitempty"`
	UpdatedAt   *int64           `json:"updatedAt,omitempty"`
}

func (this *GroupDto) FromGroup(group domain.Group) {
	model.MustCopy(group.AuditableBase, this)
	model.MustCopy(group.ModelBase, this)
	model.MustCopy(group, this)
	if group.Org != nil {
		this.Org = &GetGroupRespOrg{}
		this.Org.FromOrg(group.Org)
	}
}

// TODO: Replace with OrganizationDto
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

type CreateGroupRequest = it.CreateGroupCommand
type CreateGroupResponse = httpserver.RestCreateResponse

type DeleteGroupRequest = it.DeleteGroupCommand
type DeleteGroupResponse = httpserver.RestDeleteResponse

type UpdateGroupRequest = it.UpdateGroupCommand
type UpdateGroupResponse = httpserver.RestUpdateResponse

type GetGroupByIdRequest = it.GetGroupByIdQuery
type GetGroupByIdResponse = GroupDto

type SearchGroupsRequest = it.SearchGroupsQuery

type SearchGroupsResponse struct {
	Items []GroupDto `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

func (this *SearchGroupsResponse) FromResult(result *it.SearchGroupsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(group domain.Group) GroupDto {
		item := GroupDto{}
		item.FromGroup(group)
		return item
	})
}

type ManageUsersRequest = it.AddRemoveUsersCommand
type ManageUsersResponse = httpserver.RestUpdateResponse
