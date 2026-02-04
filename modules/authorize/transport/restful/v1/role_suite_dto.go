package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role_suite"
)

type RoleSuiteDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name                 string    `json:"name"`
	Description          *string   `json:"description,omitempty"`
	OwnerType            string    `json:"ownerType"`
	OwnerRef             string    `json:"ownerRef"`
	IsRequestable        *bool     `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool     `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool     `json:"isRequiredComment,omitempty"`
	CreatedBy            model.Id  `json:"createdBy"`
	// OrgId                *model.Id `json:"orgId,omitempty"`

	Roles        []RoleSummaryDto        `json:"roles,omitempty"`
	Organization *OrganizationSummaryDto `json:"org,omitempty"`
}

type RoleSuiteSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *RoleSuiteDto) FromRoleSuite(roleSuite domain.RoleSuite) {
	model.MustCopy(roleSuite.AuditableBase, this)
	model.MustCopy(roleSuite.ModelBase, this)
	model.MustCopy(roleSuite, this)

	this.Roles = array.Map(roleSuite.Roles, func(role domain.Role) RoleSummaryDto {
		item := RoleSummaryDto{}
		item.FromRole(&role)
		return item
	})

	// Combine OrgId and OrgName into Organization object
	if roleSuite.OrgId != nil {
		this.Organization = &OrganizationSummaryDto{}
		this.Organization.FromOrganization(roleSuite.OrgId, roleSuite.OrgName)
	}
}

type CreateRoleSuiteRequest = it.CreateRoleSuiteCommand
type CreateRoleSuiteResponse = httpserver.RestCreateResponse

type UpdateRoleSuiteRequest = it.UpdateRoleSuiteCommand
type UpdateRoleSuiteResponse = httpserver.RestUpdateResponse

type DeleteRoleSuiteRequest = it.DeleteRoleSuiteCommand
type DeleteRoleSuiteResponse = httpserver.RestDeleteResponse

type GetRoleSuiteByIdRequest = it.GetRoleSuiteByIdQuery
type GetRoleSuiteByIdResponse = RoleSuiteDto

type SearchRoleSuitesRequest = it.SearchRoleSuitesCommand
type SearchRoleSuitesResponse httpserver.RestSearchResponse[RoleSuiteDto]

func (this *SearchRoleSuitesResponse) FromResult(result *it.SearchRoleSuitesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(roleSuite domain.RoleSuite) RoleSuiteDto {
		item := RoleSuiteDto{}
		item.FromRoleSuite(roleSuite)
		return item
	})
}
