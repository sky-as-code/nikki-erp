package organization

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateOrganizationCommand)(nil)
	req = (*UpdateOrganizationCommand)(nil)
	req = (*DeleteOrganizationCommand)(nil)
	req = (*GetOrganizationBySlugQuery)(nil)
	req = (*SearchOrganizationsQuery)(nil)
	util.Unused(req)
}

var createOrganizationCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "create",
}

type CreateOrganizationCommand struct {
	Address     *string    `json:"address,omitempty"`
	DisplayName *string    `json:"displayName"`
	LegalName   *string    `json:"legalName,omitempty"`
	PhoneNumber *string    `json:"phoneNumber,omitempty"`
	Slug        model.Slug `json:"slug"`
}

func (CreateOrganizationCommand) CqrsRequestType() cqrs.RequestType {
	return createOrganizationCommandType
}

type CreateOrganizationResult model.OpResult[*domain.Organization]

var updateOrganizationCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "update",
}

type UpdateOrganizationCommand struct {
	Slug model.Slug `param:"slug" json:"slug"`

	Address     *string     `json:"address"`
	DisplayName *string     `json:"displayName"`
	Etag        model.Etag  `json:"etag"`
	LegalName   *string     `json:"legalName"`
	PhoneNumber *string     `json:"phoneNumber"`
	NewSlug     *model.Slug `json:"newSlug"`
	Status      *string     `json:"status" model:"-"`
}

func (UpdateOrganizationCommand) CqrsRequestType() cqrs.RequestType {
	return updateOrganizationCommandType
}

func (this UpdateOrganizationCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.EtagValidateRule(&this.Etag, true),
		model.SlugValidateRule(&this.Slug, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateOrganizationResult model.OpResult[*domain.Organization]

var deleteOrganizationCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "delete",
}

type DeleteOrganizationCommand struct {
	Slug model.Slug `param:"slug" json:"slug"`
}

func (DeleteOrganizationCommand) CqrsRequestType() cqrs.RequestType {
	return deleteOrganizationCommandType
}

func (this DeleteOrganizationCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.SlugValidateRule(&this.Slug, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteOrganizationResultData struct {
	Id        model.Id  `json:"id"`
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteOrganizationResult model.OpResult[*DeleteOrganizationResultData]

var getOrganizationByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "getOrganizationById",
}

type GetOrganizationByIdResult model.OpResult[*domain.Organization]

var getOrganizationBySlugQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "getOrganizationBySlug",
}

type GetOrganizationBySlugQuery struct {
	Slug           model.Slug `param:"slug" json:"slug"`
	IncludeDeleted bool       `query:"includeDeleted" json:"includeDeleted,omitempty"`
}

func (GetOrganizationBySlugQuery) CqrsRequestType() cqrs.RequestType {
	return getOrganizationBySlugQueryType
}

func (this GetOrganizationBySlugQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.SlugValidateRule(&this.Slug, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetOrganizationBySlugResult model.OpResult[*domain.Organization]

var searchOrganizationsQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "organization",
	Action:    "search",
}

type SearchOrganizationsQuery struct {
	Page           *int    `json:"page" query:"page"`
	Size           *int    `json:"size" query:"size"`
	Graph          *string `json:"graph" query:"graph"`
	IncludeDeleted bool    `json:"includeDeleted" query:"includeDeleted"`
}

func (SearchOrganizationsQuery) CqrsRequestType() cqrs.RequestType {
	return searchOrganizationsQueryType
}

func (this *SearchOrganizationsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchOrganizationsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.PageIndexValidateRule(&this.Page),
		model.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchOrganizationsResultData = crud.PagedResult[domain.Organization]
type SearchOrganizationsResult model.OpResult[*SearchOrganizationsResultData]
