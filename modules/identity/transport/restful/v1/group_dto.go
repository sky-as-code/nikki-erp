package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/group"
)

type CreateGroupRequest = it.CreateGroupCommand

type CreateGroupResponse struct {
	Id          model.Id   `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Etag        model.Etag `json:"etag"`
	OrgId       *model.Id  `json:"orgId,omitempty"`
}

func (this *CreateGroupResponse) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
	this.Description = group.Description
	this.Etag = *group.Etag
	this.OrgId = group.OrgId
}

type DeleteGroupRequest = it.DeleteGroupCommand

type DeleteGroupResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

type UpdateGroupRequest = it.UpdateGroupCommand

type UpdateGroupResponse struct {
	Id          model.Id   `param:"id" json:"id"`
	Name        string     `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Etag        model.Etag `json:"etag,omitempty"`
	OrgId       *model.Id  `json:"orgId,omitempty"`
}

func (this *UpdateGroupResponse) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
	this.Etag = *group.Etag
	this.Description = group.Description
	this.OrgId = group.OrgId
}

type GetGroupByIdRequest = it.GetGroupByIdQuery

type GetGroupRespWithOrg struct {
	Id          model.Id `json:"Id"`
	DisplayName string   `json:"displayName"`
	Slug        string   `json:"slug"`
}

type GetGroupByIdResponse struct {
	Id          model.Id             `json:"id"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	Etag        model.Etag           `json:"etag"`
	Org         *GetGroupRespWithOrg `json:"org,omitempty"`
}

func (this *GetGroupByIdResponse) FromGroup(group domain.Group) {
	this.Id = *group.Id
	this.Name = *group.Name
	this.Description = group.Description
	this.Etag = *group.Etag
	if group.Org != nil {
		this.Org.Id = *group.OrgId
		this.Org.DisplayName = *group.Org.DisplayName
		this.Org.Slug = group.Org.Slug.String()
	}
}
