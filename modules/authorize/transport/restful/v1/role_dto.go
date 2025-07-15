package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	"github.com/thoas/go-funk"
)

type CreateRoleRequest = it.CreateRoleCommand
type CreateRoleResponse = GetRoleByIdResponse

// type UpdateResourceRequest = it.UpdateResourceCommand
// type UpdateResourceResponse = GetResourceByIdResponse

type GetRoleByIdRequest = it.GetRoleByIdQuery

type GetRoleByIdResponse struct {
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

func (this *GetRoleByIdResponse) FromRole(role domain.Role) {
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

type SearchRolesRequest = it.SearchRolesQuery

type Entitlement struct {
	Id model.Id `json:"id"`

	Name *string `json:"name,omitempty"`
}

func (this *Entitlement) FromEntitlement(entitlement domain.Entitlement) {
	this.Id = *entitlement.Id
	this.Name = entitlement.Name
}

type SearchRolesResponseItem struct {
	Id model.Id `json:"id,omitempty"`

	Name                 string   `json:"name,omitempty"`
	OwnerType            string   `json:"ownerType,omitempty"`
	OwnerRef             model.Id `json:"ownerRef,omitempty"`
	IsRequestable        bool     `json:"isRequestable,omitempty"`
	IsRequiredAttachment bool     `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    bool     `json:"isRequiredComment,omitempty"`
	CreatedBy            model.Id `json:"createdBy,omitempty"`

	Entitlements []*Entitlement `json:"entitlements,omitempty"`
}

func (this *SearchRolesResponseItem) FromRole(role domain.Role) {
	this.Id = *role.Id
	this.Name = *role.Name
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

type SearchRolesResponse struct {
	Items []SearchRolesResponseItem `json:"items"`
	Total int                       `json:"total"`
	Page  int                       `json:"page"`
	Size  int                       `json:"size"`
}

func (this *SearchRolesResponse) FromResult(result *it.SearchRolesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(role *domain.Role) SearchRolesResponseItem {
		item := SearchRolesResponseItem{}
		item.FromRole(*role)
		return item
	}).([]SearchRolesResponseItem)
}
