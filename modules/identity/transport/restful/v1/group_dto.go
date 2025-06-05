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
	OrgId       string  `json:"orgId"`
}

func (this *CreateGroupResponse) FromGroup(group domain.Group) {
	this.Id = group.Id.String()
	this.Name = *group.Name
	this.Description = group.Description
	this.Etag = group.Etag.String()
	this.OrgId = group.OrgId.String()
}

type DeleteGroupRequest = it.DeleteGroupCommand

type DeleteGroupResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

type UpdateGroupRequest = it.UpdateGroupCommand

type UpdateGroupResponse struct {
	Id          string `param:"id" json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Etag        string `json:"etag,omitempty"`
	OrgId       string `json:"orgId,omitempty"`
}

func (this *UpdateGroupResponse) FromGroup(group domain.Group) {
	this.Id = group.Id.String()
	this.Name = *group.Name
	this.Etag = group.Etag.String()
	this.Description = *group.Description
	this.OrgId = group.OrgId.String()
}

type GetGroupByIdRequest = it.GetGroupByIdQuery

type GetGroupByIdResponse struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Etag        string  `json:"etag"`
	OrgId       string  `json:"orgId,omitempty"`
	DisplayName string  `json:"displayName,omitempty"`
	Slug        string  `json:"slug,omitempty"`
}

func (this *GetGroupByIdResponse) FromGroup(group domain.GroupWithOrg) {
	this.Id = group.Group.Id.String()
	this.Name = *group.Group.Name
	this.Description = group.Group.Description
	this.Etag = group.Group.Etag.String()
	if group.Organization != nil {
		this.OrgId = group.Organization.Id.String()
		this.DisplayName = *group.Organization.DisplayName
		this.Slug = group.Organization.Slug.String()
	} else {
		this.OrgId = ""
		this.DisplayName = ""
		this.Slug = ""
	}
}
