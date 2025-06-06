package v1

import (
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type CreateGroupRequest = it.CreateGroupCommand

type CreateGroupResponse struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Etag        string  `json:"etag"`
	OrgId       *string `json:"orgId,omitempty"`
}

func (this *CreateGroupResponse) FromGroup(group domain.GroupWithOrg) {
	this.Id = group.Group.Id.String()
	this.Name = group.Group.Name
	this.Description = group.Group.Description
	this.Etag = group.Group.Etag.String()
	this.OrgId = group.Group.OrgId
}

type DeleteGroupRequest = it.DeleteGroupCommand

type DeleteGroupResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

type UpdateGroupRequest = it.UpdateGroupCommand

type UpdateGroupResponse struct {
	Id          string  `param:"id" json:"id"`
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Etag        string  `json:"etag,omitempty"`
	OrgId       *string `json:"orgId,omitempty"`
}

func (this *UpdateGroupResponse) FromGroup(group domain.GroupWithOrg) {
	this.Id = group.Group.Id.String()
	this.Name = group.Group.Name
	this.Etag = group.Group.Etag.String()
	this.Description = group.Group.Description
	this.OrgId = group.Group.OrgId
}

type GetGroupByIdRequest = it.GetGroupByIdQuery

type OrganizationResponseWithGroup struct {
	Id          string `json:"Id"`
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

type GetGroupByIdResponse struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Etag        string  `json:"etag"`
	Org         OrganizationResponseWithGroup
}

func (this *GetGroupByIdResponse) FromGroup(group domain.GroupWithOrg) {
	this.Id = group.Group.Id.String()
	this.Name = group.Group.Name
	this.Description = group.Group.Description
	this.Etag = group.Group.Etag.String()
	if group.Organization != nil {
		this.Org.Id = *group.Group.OrgId
		this.Org.DisplayName = *group.Organization.DisplayName
		this.Org.Slug = group.Organization.Slug.String()
	}
}
