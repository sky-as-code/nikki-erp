package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	"github.com/thoas/go-funk"
)

type CreateResourceRequest = it.CreateResourceCommand
type CreateResourceResponse = ResponseResourceItem

type UpdateResourceRequest = it.UpdateResourceCommand
type UpdateResourceResponse = ResponseResourceItem

type GetResourceByNameRequest = it.GetResourceByNameQuery
type GetResourceByNameResponse = ResponseResourceItem

type Action struct {
	Id model.Id `json:"id"`

	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (this *Action) FromAction(action domain.Action) {
	this.Id = *action.ModelBase.Id
	this.Name = *action.Name
	this.Description = action.Description
}

type SearchResourcesRequest = it.SearchResourcesQuery

type ResponseResourceItem struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ResourceType string  `json:"resourceType"`
	ResourceRef  string  `json:"resourceRef"`
	ScopeType    string  `json:"scopeType"`

	Actions []Action `json:"actions,omitempty"`
}

func (this *ResponseResourceItem) FromResource(resource domain.Resource) {
	this.Id = *resource.Id
	this.Etag = *resource.Etag
	this.Name = *resource.Name
	this.Description = resource.Description
	this.ResourceType = resource.ResourceType.String()
	this.ResourceRef = *resource.ResourceRef
	this.ScopeType = resource.ScopeType.String()

	// Convert actions to Action array
	this.Actions = array.Map(resource.Actions, func(action domain.Action) Action {
		actionItem := Action{}
		actionItem.FromAction(action)
		return actionItem
	})
}

type SearchResourcesResponse struct {
	Items []ResponseResourceItem `json:"items"`
	Total int                    `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

func (this *SearchResourcesResponse) FromResult(result *it.SearchResourcesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(resource domain.Resource) ResponseResourceItem {
		item := ResponseResourceItem{}
		item.FromResource(resource)
		return item
	}).([]ResponseResourceItem)
}
