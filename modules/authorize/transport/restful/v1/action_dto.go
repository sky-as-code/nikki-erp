package v1

import (
	"github.com/thoas/go-funk"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
)

type ActionDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	ResourceId  model.Id `json:"resourceId"`
	CreatedBy   string   `json:"createdBy"`

	Resource *Resource `json:"resource,omitempty"`
}

type Resource struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

type CreateActionRequest = it.CreateActionCommand
type CreateActionResponse = httpserver.RestCreateResponse

type UpdateActionRequest = it.UpdateActionCommand
type UpdateActionResponse = httpserver.RestUpdateResponse

type GetActionByIdRequest = it.GetActionByIdQuery
type GetActionByIdResponse = ActionDto

type SearchActionsRequest = it.SearchActionsCommand
type SearchActionsResponse httpserver.RestSearchResponse[ActionDto]

func (this *SearchActionsResponse) FromResult(result *it.SearchActionsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(action domain.Action) ActionDto {
		item := ActionDto{}
		item.FromAction(action)
		return item
	}).([]ActionDto)
}

func (this *ActionDto) FromAction(action domain.Action) {
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
