package group

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateGroupCommand)(nil)
	req = (*UpdateGroupCommand)(nil)
	req = (*DeleteGroupCommand)(nil)
	req = (*GetGroupByIdQuery)(nil)
	util.Unused(req)
}

var createGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "create",
}

type CreateGroupCommand struct {
	CreatedBy   string `json:"createdBy"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OrgId       string `json:"orgId"`
}

func (CreateGroupCommand) Type() cqrs.RequestType {
	return createGroupCommandType
}

type CreateGroupResult model.OpResult[*domain.Group]

var updateGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "update",
}

type UpdateGroupCommand struct {
	Id          string `param:"id" json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Etag        string `json:"etag,omitempty"`
	OrgId       string `json:"orgId,omitempty"`
	UpdatedBy   string `json:"updatedBy,omitempty"`
}

func (UpdateGroupCommand) Type() cqrs.RequestType {
	return updateGroupCommandType
}

type UpdateGroupResult model.OpResult[*domain.Group]

var deleteGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "delete",
}

type DeleteGroupCommand struct {
	Id        string `json:"id" param:"id"`
	DeletedBy string `json:"deletedBy"`
}

func (DeleteGroupCommand) Type() cqrs.RequestType {
	return deleteGroupCommandType
}

type DeleteGroupResultData struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteGroupResult model.OpResult[DeleteGroupResultData]

var getGroupByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "getGroupById",
}

type GetGroupByIdQuery struct {
	Id      string `param:"id" json:"id"`
	WithOrg bool   `query:"withOrg" json:"withOrg,omitempty"`
}

func (GetGroupByIdQuery) Type() cqrs.RequestType {
	return getGroupByIdQueryType
}

type GetGroupByIdResult model.OpResult[*domain.GroupWithOrg]
