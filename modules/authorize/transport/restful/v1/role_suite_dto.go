package v1

// import (
// 	"github.com/sky-as-code/nikki-erp/common/array"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
// 	"github.com/thoas/go-funk"
// )

// type CreateResourceRequest = it.CreateResourceCommand
// type CreateResourceResponse = GetResourceByIdResponse

// type UpdateResourceRequest = it.UpdateResourceCommand
// type UpdateResourceResponse = GetResourceByIdResponse

// type GetResourceByNameRequest = it.GetResourceByNameCommand
// type GetResourceByNameResponse = GetResourceByIdResponse

// type GetResourceByIdResponse struct {
// 	Id           model.Id   `json:"id,omitempty"`
// 	Name         string     `json:"name,omitempty"`
// 	Description  string     `json:"description,omitempty"`
// 	Etag         model.Etag `json:"etag,omitempty"`
// 	ResourceType string     `json:"resourceType,omitempty"`
// 	ResourceRef  string     `json:"resourceRef,omitempty"`
// 	ScopeType    string     `json:"scopeType,omitempty"`
// }

// func (this *GetResourceByIdResponse) FromResource(resource domain.Resource) {
// 	this.Id = *resource.Id
// 	this.Name = *resource.Name
// 	this.Description = *resource.Description
// 	this.Etag = *resource.Etag
// 	this.ResourceType = resource.ResourceType.String()
// 	this.ResourceRef = *resource.ResourceRef
// 	this.ScopeType = resource.ScopeType.String()
// }

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
