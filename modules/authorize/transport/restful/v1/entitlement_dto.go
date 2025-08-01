package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
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

	Resource *ResourceDto `json:"resource,omitempty"`
	Action   *ActionDto   `json:"action,omitempty"`
	Subject  []Subject    `json:"subject,omitempty"`
}

type Subject struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *EntitlementDto) FromEntitlement(entitlement domain.Entitlement) {
	model.MustCopy(entitlement.AuditableBase, this)
	model.MustCopy(entitlement.ModelBase, this)
	model.MustCopy(entitlement, this)

	if entitlement.Resource != nil {
		this.Resource = &ResourceDto{}
		this.Resource.FromResource(*entitlement.Resource)
	}

	if entitlement.Action != nil {
		this.Action = &ActionDto{}
		this.Action.FromAction(*entitlement.Action)
	}
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
	this.Items = array.Map(result.Items, func(entitlement domain.Entitlement) EntitlementDto {
		item := EntitlementDto{}
		item.FromEntitlement(entitlement)
		return item
	})
}
