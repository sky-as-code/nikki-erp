package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/thoas/go-funk"
)

type EntitlementDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ActionId    *model.Id `json:"actionId,omitempty"`
	ResourceId  *model.Id `json:"resourceId,omitempty"`
	ScopeRef    *model.Id `json:"scopeRef,omitempty"`
	ActionExpr  *string   `json:"actionExpr,omitempty"`
	CreatedBy   model.Id  `json:"createdBy"`

	Resource *Resource  `json:"resource,omitempty"`
	Action   *Action    `json:"action,omitempty"`
	Subject  []*Subject `json:"subject,omitempty"`
}

type Subject struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

type CreateEntitlementRequest = it.CreateEntitlementCommand
type CreateEntitlementResponse = httpserver.RestCreateResponse

type UpdateEntitlementRequest = it.UpdateEntitlementCommand
type UpdateEntitlementResponse = httpserver.RestUpdateResponse

type GetEntitlementByIdRequest = it.GetEntitlementByIdQuery

type SearchEntitlementsRequest = it.SearchEntitlementsQuery
type SearchEntitlementsResponse httpserver.RestSearchResponse[EntitlementDto]

func (this *SearchEntitlementsResponse) FromResult(result *it.SearchEntitlementsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(entitlement *domain.Entitlement) EntitlementDto {
		item := EntitlementDto{}
		item.FromEntitlement(*entitlement)
		return item
	}).([]EntitlementDto)
}

func (this *EntitlementDto) FromEntitlement(entitlement domain.Entitlement) {
	this.Id = *entitlement.Id
	this.Etag = *entitlement.Etag
	this.Name = *entitlement.Name
	this.Description = entitlement.Description
	this.ResourceId = entitlement.ResourceId
	this.ActionId = entitlement.ActionId
	this.ScopeRef = entitlement.ScopeRef
	this.ActionExpr = entitlement.ActionExpr
	this.CreatedBy = *entitlement.CreatedBy

	if entitlement.Resource != nil {
		this.Resource = &Resource{
			Id:   *entitlement.Resource.Id,
			Name: *entitlement.Resource.Name,
		}
	}

	if entitlement.Action != nil {
		this.Action = &Action{
			Id:   *entitlement.Action.Id,
			Name: *entitlement.Action.Name,
		}
	}
}
