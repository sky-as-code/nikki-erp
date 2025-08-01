package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
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

	Resource *ResourceDto `json:"resource,omitempty"`
}

func (this *ActionDto) FromAction(action domain.Action) {
	model.MustCopy(action.ModelBase, this)
	model.MustCopy(action.AuditableBase, this)
	model.MustCopy(action, this)

	if action.Resource != nil {
		this.Resource = &ResourceDto{}
		this.Resource.FromResource(*action.Resource)
	}
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
	this.Items = array.Map(result.Items, func(action domain.Action) ActionDto {
		item := ActionDto{}
		item.FromAction(action)
		return item
	})
}
