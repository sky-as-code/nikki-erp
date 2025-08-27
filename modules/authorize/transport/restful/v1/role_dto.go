package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

type RoleDto struct {
	Id        model.Id   `json:"id"`
	Etag      model.Etag `json:"etag"`
	CreatedAt time.Time  `json:"createdAt"`

	Name                 string   `json:"name"`
	Description          *string  `json:"description,omitempty"`
	OwnerType            string   `json:"ownerType"`
	OwnerRef             model.Id `json:"ownerRef"`
	IsRequestable        bool     `json:"isRequestable"`
	IsRequiredAttachment bool     `json:"isRequiredAttachment"`
	IsRequiredComment    bool     `json:"isRequiredComment"`
	CreatedBy            model.Id `json:"createdBy"`

	Entitlements []EntitlementSummaryDto `json:"entitlements,omitempty"`
}

type RoleSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *RoleDto) FromRole(role domain.Role) {
	model.MustCopy(role.AuditableBase, this)
	model.MustCopy(role.ModelBase, this)
	model.MustCopy(role, this)

	this.Entitlements = array.Map(role.Entitlements, func(entitlement domain.Entitlement) EntitlementSummaryDto {
		entitlementItem := EntitlementSummaryDto{}
		entitlementItem.FromEntitlement(&entitlement)
		return entitlementItem
	})
}

func (this *RoleSummaryDto) FromRole(role *domain.Role) {
	this.Id = *role.Id
	this.Name = *role.Name
}

type CreateRoleRequest = it.CreateRoleCommand
type CreateRoleResponse = httpserver.RestCreateResponse

type UpdateRoleRequest = it.UpdateRoleCommand
type UpdateRoleResponse = httpserver.RestUpdateResponse

type DeleteRoleHardRequest = it.DeleteRoleHardCommand
type DeleteRoleHardResponse = httpserver.RestDeleteResponse

type GetRoleByIdRequest = it.GetRoleByIdQuery
type GetRoleByIdResponse = RoleDto

type SearchRolesRequest = it.SearchRolesQuery
type SearchRolesResponse httpserver.RestSearchResponse[RoleDto]

func (this *SearchRolesResponse) FromResult(result *it.SearchRolesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(role domain.Role) RoleDto {
		item := RoleDto{}
		item.FromRole(role)
		return item
	})
}
