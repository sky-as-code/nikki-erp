package v1

// import (
// 	"time"

// 	"github.com/sky-as-code/nikki-erp/common/array"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role"
// )

// type RoleDto struct {
// 	Id        model.Id   `json:"id"`
// 	Etag      model.Etag `json:"etag"`
// 	CreatedAt time.Time  `json:"createdAt"`

// 	Name                 string    `json:"name"`
// 	Description          *string   `json:"description,omitempty"`
// 	OwnerType            string    `json:"ownerType"`
// 	OwnerRef             model.Id  `json:"ownerRef"`
// 	IsRequestable        bool      `json:"isRequestable"`
// 	IsRequiredAttachment bool      `json:"isRequiredAttachment"`
// 	IsRequiredComment    bool      `json:"isRequiredComment"`
// 	CreatedBy            model.Id  `json:"createdBy"`
// 	// OrgId                *model.Id `json:"orgId,omitempty"`

// 	Entitlements []EntitlementSummaryDto `json:"entitlements,omitempty"`
// 	Organization *OrganizationSummaryDto `json:"org,omitempty"`
// }

// type RoleSummaryDto struct {
// 	Id   model.Id `json:"id"`
// 	Name string   `json:"name"`
// }

// func (this *RoleDto) FromRole(role domain.Role) {
// 	if id := role.GetId(); id != nil {
// 		this.Id = *id
// 	}
// 	if e := role.GetEtag(); e != nil {
// 		this.Etag = *e
// 	}
// 	if t := role.GetCreatedAt(); t != nil {
// 		this.CreatedAt = *t
// 	}
// 	if n := role.GetName(); n != nil {
// 		this.Name = *n
// 	}
// 	this.Description = role.GetDescription()
// 	if ot := role.GetOwnerType(); ot != nil {
// 		this.OwnerType = string(*ot)
// 	}
// 	if ref := role.GetOwnerRef(); ref != nil {
// 		this.OwnerRef = *ref
// 	}
// 	if v := role.GetIsRequestable(); v != nil {
// 		this.IsRequestable = *v
// 	}
// 	if v := role.GetIsRequiredAttachment(); v != nil {
// 		this.IsRequiredAttachment = *v
// 	}
// 	if v := role.GetIsRequiredComment(); v != nil {
// 		this.IsRequiredComment = *v
// 	}
// 	if cb := role.GetCreatedBy(); cb != nil {
// 		this.CreatedBy = *cb
// 	}

// 	this.Entitlements = array.Map(role.Entitlements, func(entitlement domain.Entitlement) EntitlementSummaryDto {
// 		entitlementItem := EntitlementSummaryDto{}

// 		var scopeRefId *model.Id
// 		if entitlement.ScopeRef != nil {
// 			id := model.Id(*entitlement.ScopeRef)
// 			scopeRefId = &id
// 		}

// 		entitlementItem.FromEntitlementWithScopeRef(&entitlement, scopeRefId)
// 		return entitlementItem
// 	})

// 	// Combine OrgId and OrgName into Organization object
// 	if oid := role.GetOrgId(); oid != nil {
// 		this.Organization = &OrganizationSummaryDto{}
// 		this.Organization.FromOrganization(oid, role.OrgName)
// 	}
// }

// func (this *RoleSummaryDto) FromRole(role *domain.Role) {
// 	if id := role.GetId(); id != nil {
// 		this.Id = *id
// 	}
// 	if n := role.GetName(); n != nil {
// 		this.Name = *n
// 	}
// }

// type AddEntitlementsRequest = it.AddEntitlementsCommand
// type AddEntitlementsResponse = httpserver.RestUpdateResponse

// type RemoveEntitlementsRequest = it.RemoveEntitlementsCommand
// type RemoveEntitlementsResponse = httpserver.RestUpdateResponse

// type CreateRoleRequest = it.CreateRoleCommand
// type CreateRoleResponse = httpserver.RestCreateResponse

// type UpdateRoleRequest = it.UpdateRoleCommand
// type UpdateRoleResponse = httpserver.RestUpdateResponse

// type DeleteRoleHardRequest = it.DeleteRoleHardCommand
// type DeleteRoleHardResponse = httpserver.RestDeleteResponse

// type GetRoleByIdRequest = it.GetRoleByIdQuery
// type GetRoleByIdResponse = RoleDto

// type SearchRolesRequest = it.SearchRolesQuery
// type SearchRolesResponse httpserver.RestSearchResponse[RoleDto]

// func (this *SearchRolesResponse) FromResult(result *it.SearchRolesResultData) {
// 	this.Total = result.Total
// 	this.Page = result.Page
// 	this.Size = result.Size
// 	this.Items = array.Map(result.Items, func(role domain.Role) RoleDto {
// 		item := RoleDto{}
// 		item.FromRole(role)
// 		return item
// 	})
// }
