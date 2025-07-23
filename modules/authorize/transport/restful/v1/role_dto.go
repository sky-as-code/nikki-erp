package v1

import (
	"time"

	"github.com/thoas/go-funk"

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

	Entitlements []*Entitlement `json:"entitlements,omitempty"`
}

type Entitlement struct {
	Id model.Id `json:"id"`

	Name *string `json:"name,omitempty"`
}

type CreateRoleRequest = it.CreateRoleCommand
type CreateRoleResponse = httpserver.RestCreateResponse

type GetRoleByIdRequest = it.GetRoleByIdQuery
type GetRoleByIdResponse = RoleDto

type SearchRolesRequest = it.SearchRolesQuery
type SearchRolesResponse httpserver.RestSearchResponse[RoleDto]

func (this *SearchRolesResponse) FromResult(result *it.SearchRolesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(role *domain.Role) RoleDto {
		item := RoleDto{}
		item.FromRole(*role)
		return item
	}).([]RoleDto)
}

func (this *RoleDto) FromRole(role domain.Role) {
	this.Id = *role.Id
	this.Etag = *role.Etag
	this.CreatedAt = *role.CreatedAt
	this.Name = *role.Name
	this.Description = role.Description
	this.OwnerType = role.OwnerType.String()
	this.OwnerRef = *role.OwnerRef
	this.IsRequestable = *role.IsRequestable
	this.IsRequiredAttachment = *role.IsRequiredAttachment
	this.IsRequiredComment = *role.IsRequiredComment
	this.CreatedBy = *role.CreatedBy

	this.Entitlements = array.Map(role.Entitlements, func(entitlement *domain.Entitlement) *Entitlement {
		entitlementItem := &Entitlement{}
		entitlementItem.FromEntitlement(*entitlement)
		return entitlementItem
	})
}

func (this *Entitlement) FromEntitlement(entitlement domain.Entitlement) {
	this.Id = *entitlement.Id
	this.Name = entitlement.Name
}
