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
	ActionExpr  *string   `json:"actionExpr,omitempty"`
	CreatedBy   model.Id  `json:"createdBy"`
	OrgId       *model.Id `json:"orgId,omitempty"`

	Resource *ResourceSummaryDto `json:"resource,omitempty"`
	Action   *ActionSummaryDto   `json:"action,omitempty"`
	Subject  []Subject           `json:"subject,omitempty"`
}

type EntitlementSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
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
		this.Resource = &ResourceSummaryDto{}
		this.Resource.FromResource(*entitlement.Resource)
	}

	if entitlement.Action != nil {
		this.Action = &ActionSummaryDto{}
		this.Action.FromAction(*entitlement.Action)
	}
}

func (this *EntitlementSummaryDto) FromEntitlement(entitlement *domain.Entitlement) {
	this.Id = *entitlement.Id
	this.Name = *entitlement.Name
}

type CreateEntitlementRequest = it.CreateEntitlementCommand
type CreateEntitlementResponse = httpserver.RestCreateResponse

type UpdateEntitlementRequest = it.UpdateEntitlementCommand
type UpdateEntitlementResponse = httpserver.RestUpdateResponse

type DeleteEntitlementHardByIdRequest = it.DeleteEntitlementHardByIdCommand
type DeleteEntitlementHardByIdResponse = httpserver.RestDeleteResponse

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
