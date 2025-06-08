package group

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
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
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OrgId       *model.Id `json:"orgId,omitempty"`
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
	Id          model.Id   `param:"id" json:"id"`
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	Etag        model.Etag `json:"etag"`
	OrgId       *model.Id  `json:"orgId,omitempty"`
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
	Id model.Id `json:"id" param:"id"`
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
	Id      model.Id `param:"id" json:"id"`
	WithOrg *bool    `query:"withOrg" json:"withOrg,omitempty"`
}

func (GetGroupByIdQuery) Type() cqrs.RequestType {
	return getGroupByIdQueryType
}

func (this *GetGroupByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type GetGroupByIdResult model.OpResult[*domain.Group]
