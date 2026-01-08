package hierarchy

import (
	"time"

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
	req = (*CreateHierarchyLevelCommand)(nil)
	req = (*UpdateHierarchyLevelCommand)(nil)
	req = (*DeleteHierarchyLevelCommand)(nil)
	req = (*GetHierarchyLevelByIdQuery)(nil)
	req = (*SearchHierarchyLevelsQuery)(nil)
	req = (*ExistsHierarchyLevelByIdQuery)(nil)
	util.Unused(req)
}

// CreateHierarchyLevelCommand
var createHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "create",
}

type CreateHierarchyLevelCommand struct {
	Name     string    `json:"name"`
	OrgId    model.Id  `json:"orgId"`
	ParentId *model.Id `json:"parentId"`
}

func (CreateHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return createHierarchyLevelCommandType
}

type CreateHierarchyLevelResult = crud.OpResult[*domain.HierarchyLevel]

// UpdateHierarchyLevelCommand
var updateHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "update",
}

type UpdateHierarchyLevelCommand struct {
	Id       model.Id   `param:"id" json:"id"`
	Name     string     `json:"name"`
	OrgId    *model.Id  `json:"orgId"`
	ParentId *model.Id  `json:"parentId"`
	Etag     model.Etag `json:"etag"`
}

func (UpdateHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return updateHierarchyLevelCommandType
}

type UpdateHierarchyLevelResult = crud.OpResult[*domain.HierarchyLevel]

// DeleteHierarchyLevelCommand
var deleteHierarchyLevelCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "delete",
}

type DeleteHierarchyLevelCommand struct {
	Id model.Id `param:"id" json:"id"`
}

type DeleteHierarchyrResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (DeleteHierarchyLevelCommand) CqrsRequestType() cqrs.RequestType {
	return deleteHierarchyLevelCommandType
}

func (this DeleteHierarchyLevelCommand) ToDomainModel() *domain.HierarchyLevel {
	hier := &domain.HierarchyLevel{}
	hier.Id = &this.Id
	return hier
}

func (this DeleteHierarchyLevelCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteHierarchyLevelResult = crud.DeletionResult

// GetHierarchyLevelByIdQuery
var getHierarchyLevelByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "getById",
}

type GetHierarchyLevelByIdQuery struct {
	Id             model.Id `param:"id" json:"id"`
	IncludeDeleted bool     `query:"includeDeleted" json:"includeDeleted"`
	WithChildren   bool     `query:"withChildren" json:"withChildren"`
}

func (GetHierarchyLevelByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getHierarchyLevelByIdQueryType
}

func (this GetHierarchyLevelByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetHierarchyLevelByIdResult = crud.OpResult[*domain.HierarchyLevel]

// SearchHierarchyLevelsQuery
var searchHierarchyLevelsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "search",
}

type SearchHierarchyLevelsQuery struct {
	crud.SearchQuery
	IncludeDeleted bool `query:"includeDeleted" json:"includeDeleted"`
	WithChildren   bool `query:"withChildren" json:"withChildren"`
	WithOrg        bool `query:"withOrg" json:"withOrg"`
	WithParent     bool `query:"withParent" json:"withParent"`
}

func (SearchHierarchyLevelsQuery) CqrsRequestType() cqrs.RequestType {
	return searchHierarchyLevelsQueryType
}

func (this *SearchHierarchyLevelsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchHierarchyLevelsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchHierarchyLevelsResultData = crud.PagedResult[domain.HierarchyLevel]
type SearchHierarchyLevelsResult = crud.OpResult[*SearchHierarchyLevelsResultData]

// AddRemoveUsersCommand
type AddRemoveUsersCommand struct {
	HierarchyId model.Id   `param:"hierarchyId" json:"hierarchyId"`
	Add         []model.Id `json:"add"`
	Remove      []model.Id `json:"remove"`
	Etag        model.Etag `json:"etag"`
}

var addRemoveUsersCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "addRemoveUsers",
}

func (AddRemoveUsersCommand) CqrsRequestType() cqrs.RequestType {
	return addRemoveUsersCommandType
}

func (this *AddRemoveUsersCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.HierarchyId, true),
		model.IdValidateRuleMulti(&this.Add, false, 0, model.MODEL_RULE_ID_ARR_MAX),
		model.IdValidateRuleMulti(&this.Remove, false, 0, model.MODEL_RULE_ID_ARR_MAX),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type AddRemoveUsersResultData struct {
	Id        model.Id   `json:"id"`
	Etag      model.Etag `json:"etag"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type AddRemoveUsersResult = crud.OpResult[*AddRemoveUsersResultData]

var existsHierarchyLevelByIdQuery = cqrs.RequestType{
	Module:    "identity",
	Submodule: "hierarchy",
	Action:    "existsHierarchyLevelById",
}

type ExistsHierarchyLevelByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (ExistsHierarchyLevelByIdQuery) CqrsRequestType() cqrs.RequestType {
	return existsHierarchyLevelByIdQuery
}

func (this ExistsHierarchyLevelByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ExistsHierarchyLevelByIdResult = crud.OpResult[bool]
