package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type OrganizationDto struct {
	Id          string  `json:"id"`
	Address     *string `json:"address,omitempty"`
	CreatedAt   int64   `json:"createdAt,omitempty"`
	DisplayName string  `json:"displayName"`
	Etag        string  `json:"etag"`
	LegalName   *string `json:"legalName,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Slug        string  `json:"slug"`
	Status      string  `json:"status"`
	UpdatedAt   *int64  `json:"updatedAt,omitempty"`
}

func (this *OrganizationDto) FromOrg(org domain.Organization) {
	model.MustCopy(org.AuditableBase, this)
	model.MustCopy(org.ModelBase, this)
	model.MustCopy(org, this)
}

type CreateOrganizationRequest = itOrg.CreateOrganizationCommand
type CreateOrganizationResponse = httpserver.RestCreateResponse

type UpdateOrganizationRequest = itOrg.UpdateOrganizationCommand
type UpdateOrganizationResponse = httpserver.RestUpdateResponse

type DeleteOrganizationRequest = itOrg.DeleteOrganizationCommand
type DeleteOrganizationResponse = httpserver.RestDeleteResponse

type GetOrganizationBySlugRequest = itOrg.GetOrganizationBySlugQuery
type GetOrganizationBySlugResponse = OrganizationDto

type SearchOrganizationsRequest = itOrg.SearchOrganizationsQuery

type SearchOrganizationsResponse httpserver.RestSearchResponse[OrganizationDto]

func (this *SearchOrganizationsResponse) FromResult(result *itOrg.SearchOrganizationsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(org domain.Organization) OrganizationDto {
		orgDto := OrganizationDto{}
		orgDto.FromOrg(org)
		return orgDto
	})
}
