package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
)

type ActionDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	ResourceId  model.Id `json:"resourceId"`
	CreatedBy   string   `json:"createdBy"`

	Resource *ResourceSummaryDto `json:"resource,omitempty"`
}

type ActionSummaryDto struct {
	Id   model.Id `json:"id"`
	Name string   `json:"name"`
}

func (this *ActionDto) FromAction(action domain.Action) {
	model.MustCopy(action.ModelBase, this)
	model.MustCopy(action.AuditableBase, this)
	model.MustCopy(action, this)

	if action.Resource != nil {
		this.Resource = &ResourceSummaryDto{}
		this.Resource.FromResource(*action.Resource)
	}
}

func (this *ActionSummaryDto) FromAction(action domain.Action) {
	this.Id = *action.Id
	this.Name = *action.Name
}

type CreateActionRequest = it.CreateActionCommand
type CreateActionResponse = httpserver.RestCreateResponse

type UpdateActionRequest = it.UpdateActionCommand
type UpdateActionResponse = httpserver.RestUpdateResponse

type DeleteActionHardByIdRequest = it.DeleteActionHardByIdCommand
type DeleteActionHardByIdResponse = httpserver.RestDeleteResponse

type GetActionByIdRequest = it.GetActionByIdQuery
type GetActionByIdResponse = ActionDto

type SearchActionsRequest = it.SearchActionsQuery
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
