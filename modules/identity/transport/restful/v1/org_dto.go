package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	itOrg "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type OrganizationDto struct {
	Id          string     `json:"id"`
	Address     *string    `json:"address,omitempty"`
	CreatedAt   time.Time  `json:"createdAt,omitempty"`
	DisplayName string     `json:"displayName"`
	Etag        string     `json:"etag"`
	LegalName   *string    `json:"legalName,omitempty"`
	PhoneNumber *string    `json:"phoneNumber,omitempty"`
	Slug        string     `json:"slug"`
	Status      string     `json:"status"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
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
type DeleteOrganizationResponse struct {
	Slug      model.Slug `json:"slug"`
	DeletedAt int64      `json:"deletedAt"`
}

func (this *DeleteOrganizationResponse) FromNonEntity(src any) {
	model.MustCopy(src, this)
}

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

type ManageOrganizationUsersRequest = itOrg.AddRemoveUsersCommand
type ManageOrganizationUsersResponse = httpserver.RestUpdateResponse
