package v1

import (
	// "github.com/sky-as-code/nikki-erp/common/array"
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	// "github.com/thoas/go-funk"
)

type CreateRoleRequest = it.CreateRoleCommand
type CreateRoleResponse = GetRoleByIdResponse

// type UpdateResourceRequest = it.UpdateResourceCommand
// type UpdateResourceResponse = GetResourceByIdResponse

// type GetResourceByNameRequest = it.GetResourceByNameCommand
// type GetResourceByNameResponse = GetResourceByIdResponse

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
}

// type SearchResourcesRequest = it.SearchResourcesCommand

// type SearchResourcesResponseItem struct {
// 	Id           model.Id   `json:"id,omitempty"`
// 	Name         string     `json:"name,omitempty"`
// 	Description  string     `json:"description,omitempty"`
// 	Etag         model.Etag `json:"etag,omitempty"`
// 	ResourceType string     `json:"resourceType,omitempty"`
// 	ResourceRef  string     `json:"resourceRef,omitempty"`
// 	ScopeType    string     `json:"scopeType,omitempty"`
// 	Actions      []string   `json:"actions,omitempty"`
// }

// func (this *SearchResourcesResponseItem) FromResource(resource domain.Resource) {
// 	this.Id = *resource.Id
// 	this.Name = *resource.Name
// 	this.Description = *resource.Description
// 	this.Etag = *resource.Etag
// 	this.ResourceType = resource.ResourceType.String()
// 	this.ResourceRef = *resource.ResourceRef
// 	this.ScopeType = resource.ScopeType.String()

// 	// Convert actions to string array
// 	this.Actions = array.Map(resource.Actions, func(action domain.Action) string {
// 		return *action.Name
// 	})
// }

// type SearchResourcesResponse struct {
// 	Items []SearchResourcesResponseItem `json:"items"`
// 	Total int                         `json:"total"`
// 	Page  int                         `json:"page"`
// 	Size  int                         `json:"size"`
// }

// func (this *SearchResourcesResponse) FromResult(result *it.SearchResourcesResultData) {
// 	this.Total = result.Total
// 	this.Page = result.Page
// 	this.Size = result.Size
// 	this.Items = funk.Map(result.Items, func(resource domain.Resource) SearchResourcesResponseItem {
// 		item := SearchResourcesResponseItem{}
// 		item.FromResource(resource)
// 		return item
// 	}).([]SearchResourcesResponseItem)
// }
