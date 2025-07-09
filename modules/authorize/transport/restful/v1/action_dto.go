package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	"github.com/thoas/go-funk"
)

type CreateActionRequest = it.CreateActionCommand
type CreateActionResponse = ResponseActionItem

type UpdateActionRequest = it.UpdateActionCommand
type UpdateActionResponse = ResponseActionItem

type GetActionByIdRequest = it.GetActionByIdQuery
type GetActionByIdResponse = ResponseActionItem

type Resource struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

type ResponseActionItem struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ResourceId  model.Id  `json:"resourceId"`
	CreatedBy   string    `json:"createdBy"`
	Resource    *Resource `json:"resource,omitempty"`
}

func (this *ResponseActionItem) FromAction(action domain.Action) {
	this.Id = *action.Id
	this.Name = *action.Name
	this.Description = action.Description
	this.ResourceId = *action.ResourceId
	this.Etag = *action.Etag
	this.CreatedBy = *action.CreatedBy

	if action.Resource != nil {
		this.Resource = &Resource{
			Id:   *action.Resource.Id,
			Name: *action.Resource.Name,
		}
	}
}

type SearchActionsRequest = it.SearchActionsCommand
type SearchActionsResponseItem = ResponseActionItem

type SearchActionsResponse struct {
	Items []SearchActionsResponseItem `json:"items"`
	Total int                         `json:"total"`
	Page  int                         `json:"page"`
	Size  int                         `json:"size"`
}

func (this *SearchActionsResponse) FromResultWithResources(result *it.SearchActionsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(action domain.Action) SearchActionsResponseItem {
		item := SearchActionsResponseItem{}
		item.FromAction(action)
		return item
	}).([]SearchActionsResponseItem)
}
