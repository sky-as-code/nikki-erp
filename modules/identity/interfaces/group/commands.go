package group

import (
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateGroupCommand)(nil)
	req = (*UpdateGroupCommand)(nil)
	req = (*DeleteGroupCommand)(nil)
	req = (*GetGroupByIdQuery)(nil)
	req = (*GroupExistsCommand)(nil)
	util.Unused(req)
}

var addRemoveUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "addRemoveUsers",
}

type AddRemoveUsersCommand struct {
	GroupId model.Id   `param:"groupId" json:"groupId"`
	Add     []model.Id `json:"add"`
	Remove  []model.Id `json:"remove"`
	Etag    model.Etag `json:"etag"`
	ScopeRef *model.Id `query:"scopeRef" json:"scopeRef"`
}

func (AddRemoveUsersCommand) CqrsRequestType() cqrs.RequestType {
	return addRemoveUsersCommandType
}

func (this *AddRemoveUsersCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.GroupId, true),
		model.IdValidateRuleMulti(&this.Add, false, 0, model.MODEL_RULE_ID_ARR_MAX),
		model.IdValidateRuleMulti(&this.Remove, false, 0, model.MODEL_RULE_ID_ARR_MAX),
		model.IdPtrValidateRule(&this.ScopeRef, false),
		val.Field(&this.Add, val.By(func(value any) error {
			if this.Add == nil || this.Remove == nil || len(this.Remove) == 0 {
				return nil
			}
			ids, _ := value.([]model.Id)
			for _, addedId := range ids {
				if array.Contains(this.Remove, addedId) {
					return errors.New("add and remove must not contain the same id")
				}
			}
			return nil
		})),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type AddRemoveUsersResultData struct {
	Id        model.Id   `json:"id"`
	Etag      model.Etag `json:"etag"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type AddRemoveUsersResult = crud.OpResult[*AddRemoveUsersResultData]

var createGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "create",
}

type CreateGroupCommand struct {
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OrgId       *model.Id `json:"orgId"`
	ScopeRef    *model.Id `query:"scopeRef" json:"scopeRef"`
}

func (CreateGroupCommand) CqrsRequestType() cqrs.RequestType {
	return createGroupCommandType
}

type CreateGroupResult = crud.OpResult[*domain.Group]

var updateGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "update",
}

type UpdateGroupCommand struct {
	Id          model.Id   `param:"id" json:"id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Etag        model.Etag `json:"etag"`
	OrgId       *model.Id  `json:"orgId"`
	ScopeRef    *model.Id  `query:"scopeRef" json:"scopeRef"`
}

func (UpdateGroupCommand) CqrsRequestType() cqrs.RequestType {
	return updateGroupCommandType
}

type UpdateGroupResult = crud.OpResult[*domain.Group]

var deleteGroupCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "delete",
}

type DeleteGroupCommand struct {
	Id model.Id `json:"id" param:"id"`
	ScopeRef *model.Id `query:"scopeRef" json:"scopeRef"`
}

func (DeleteGroupCommand) CqrsRequestType() cqrs.RequestType {
	return deleteGroupCommandType
}

func (this DeleteGroupCommand) ToDomainModel() *domain.Group {
	group := &domain.Group{}
	group.Id = &this.Id
	return group
}

func (this DeleteGroupCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteGroupResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteGroupResult = crud.DeletionResult

var getGroupByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "getGroupById",
}

type GetGroupByIdQuery struct {
	Id      model.Id `param:"id" json:"id"`
	WithOrg bool     `query:"withOrg" json:"withOrg"`
	ScopeRef *model.Id `query:"scopeRef" json:"scopeRef"`
}

func (GetGroupByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getGroupByIdQueryType
}

func (this GetGroupByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
	// return nil
}

type GetGroupByIdResult = crud.OpResult[*domain.Group]

var searchGroupsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "search",
}

type SearchGroupsQuery struct {
	crud.SearchQuery
	WithOrg bool `json:"withOrg" query:"withOrg"`
	ScopeRef *model.Id `json:"scopeRef" query:"scopeRef"`
}

func (SearchGroupsQuery) CqrsRequestType() cqrs.RequestType {
	return searchGroupsQueryType
}

func (this *SearchGroupsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchGroupsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
		model.IdPtrValidateRule(&this.ScopeRef, false),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchGroupsResultData = crud.PagedResult[domain.Group]
type SearchGroupsResult = crud.OpResult[*SearchGroupsResultData]

var existsCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "group",
	Action:    "exists",
}

type GroupExistsCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (GroupExistsCommand) CqrsRequestType() cqrs.RequestType {
	return existsCommandType
}

func (this GroupExistsCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GroupExistsResult = crud.OpResult[bool]
