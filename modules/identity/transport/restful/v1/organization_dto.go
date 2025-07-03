package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

type CreateOrganizationRequest = it.CreateOrganizationCommand
type CreateOrganizationResponse = GetOrganizationResponse

type UpdateOrganizationRequest = it.UpdateOrganizationCommand
type UpdateOrganizationResponse = GetOrganizationResponse

type DeleteOrganizationRequest = it.DeleteOrganizationCommand

type DeleteOrganizationResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

type GetOrganizationBySlugRequest = it.GetOrganizationBySlugQuery
type GetOrganizationBySlugResponse = GetOrganizationResponse

type SearchOrganizationsRequest = it.SearchOrganizationsQuery

type SearchOrganizationsResponseItem struct {
	Id          model.Id   `json:"id"`
	DisplayName string     `json:"displayName"`
	Slug        model.Slug `json:"slug"`
	Status      string     `json:"status"`
	CreatedAt   int64      `json:"createdAt,omitempty"`
	UpdatedAt   *int64     `json:"updatedAt,omitempty"`
}

func (this *SearchOrganizationsResponseItem) FromOrganization(org domain.Organization) {
	this.Id = *org.Id
	this.DisplayName = *org.DisplayName
	this.Slug = *org.Slug
	if org.Status != nil {
		this.Status = string(*org.Status)
	}
	this.CreatedAt = org.CreatedAt.UnixMilli()
	this.UpdatedAt = safe.GetTimeUnixMilli(org.UpdatedAt)
}

type SearchOrganizationsResponse struct {
	Items []SearchOrganizationsResponseItem `json:"items"`
	Total int                               `json:"total"`
	Page  int                               `json:"page"`
	Size  int                               `json:"size"`
}

func (this *SearchOrganizationsResponse) FromResult(result *it.SearchOrganizationsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = make([]SearchOrganizationsResponseItem, len(result.Items))
	for i, org := range result.Items {
		this.Items[i].FromOrganization(org)
	}
}

type GetOrganizationResponse struct {
	Id          model.Id   `json:"id"`
	Address     *string    `json:"address,omitempty"`
	CreatedAt   int64      `json:"createdAt,omitempty"`
	DisplayName string     `json:"displayName"`
	LegalName   *string    `json:"legalName,omitempty"`
	PhoneNumber *string    `json:"phoneNumber,omitempty"`
	Slug        model.Slug `json:"slug"`
	Status      string     `json:"status"`
	Etag        model.Etag `json:"etag"`
	UpdatedAt   *int64     `json:"updatedAt,omitempty"`
}

func (this *GetOrganizationResponse) FromOrganization(org domain.Organization) {
	this.Id = *org.Id
	this.CreatedAt = org.CreatedAt.UnixMilli()
	this.DisplayName = *org.DisplayName
	this.Address = org.Address
	this.LegalName = org.LegalName
	this.PhoneNumber = org.PhoneNumber
	this.Slug = *org.Slug
	if org.Status != nil {
		this.Status = string(*org.Status)
	}
	this.Etag = *org.Etag
	this.UpdatedAt = safe.GetTimeUnixMilli(org.UpdatedAt)
}
