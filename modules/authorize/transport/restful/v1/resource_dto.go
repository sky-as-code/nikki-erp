package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
)

type ResourceDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name         string  `json:"name"`
	Description  *string `json:"description,omitempty"`
	ResourceType string  `json:"resourceType"`
	ResourceRef  string  `json:"resourceRef"`
	ScopeType    string  `json:"scopeType"`

	Actions []ActionDto `json:"actions,omitempty"`
}

func (this *ResourceDto) FromResource(resource domain.Resource) {
	model.MustCopy(resource.AuditableBase, this)
	model.MustCopy(resource.ModelBase, this)
	model.MustCopy(resource, this)

	if resource.Actions != nil {
		this.Actions = array.Map(resource.Actions, func(action domain.Action) ActionDto {
			actionDto := ActionDto{}
			actionDto.FromAction(action)
			return actionDto
		})
	}
}

type CreateResourceRequest = it.CreateResourceCommand
type CreateResourceResponse = httpserver.RestCreateResponse

type UpdateResourceRequest = it.UpdateResourceCommand
type UpdateResourceResponse = httpserver.RestUpdateResponse

type DeleteResourceHardByNameRequest = it.DeleteResourceHardByNameQuery
type DeleteResourceHardByNameResponse = httpserver.RestDeleteResponse

type GetResourceByNameRequest = it.GetResourceByNameQuery
type GetResourceByNameResponse = ResourceDto

type SearchResourcesRequest = it.SearchResourcesQuery
type SearchResourcesResponse httpserver.RestSearchResponse[ResourceDto]

func (this *SearchResourcesResponse) FromResult(result *it.SearchResourcesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(resource domain.Resource) ResourceDto {
		item := ResourceDto{}
		item.FromResource(resource)
		return item
	})
}
