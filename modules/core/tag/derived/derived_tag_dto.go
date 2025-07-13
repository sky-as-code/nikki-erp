package derived

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/core/tag/interfaces"
)

//
// NOTE: These DTOs are used for feature Tags inheriting from core Tags.
//

type CreateDerivedTagRequest = it.CreateTagCommand
type CreateDerivedTagResponse = GetDerivedTagByIdResponse

type UpdateDerivedTagRequest = it.UpdateTagCommand
type UpdateDerivedTagResponse = GetDerivedTagByIdResponse

type DeleteDerivedTagRequest = it.DeleteTagCommand
type DeleteDerivedTagResultData struct {
	DeletedAt int64 `json:"deletedAt"`
}
type DeleteDerivedTagResponse struct {
	DeletedAt int64 `json:"deletedAt"`
}

func NewDeleteDerivedTagResponse(result it.DeleteTagResult) DeleteDerivedTagResponse {
	return DeleteDerivedTagResponse{
		DeletedAt: result.Data.DeletedAt.UnixMilli(),
	}
}

type DerivedTagDto struct {
	Id    model.Id       `json:"id,omitempty"`
	Etag  model.Etag     `json:"etag,omitempty"`
	Label model.LangJson `json:"label,omitempty"`
}

func NewDerivedTagDto(src it.Tag) DerivedTagDto {
	return DerivedTagDto{
		Id:    *src.Id,
		Etag:  *src.Etag,
		Label: *src.Label,
	}
}

func NewDerivedTagDtos(src []it.Tag) []DerivedTagDto {
	return array.Map(src, func(item it.Tag) DerivedTagDto {
		return NewDerivedTagDto(item)
	})
}

type GetDerivedTagByIdRequest = it.GetTagByIdQuery
type GetDerivedTagByIdResponse = DerivedTagDto

type SearchDerivedTagsRequest = it.ListTagsQuery
type SearchDerivedTagsResponse = crud.PagedResult[DerivedTagDto]

type ListDerivedTagsRequest = it.ListTagsQuery
type ListDerivedTagsResponse = crud.PagedResult[DerivedTagDto]

func NewListDerivedTagsResponse(src it.ListTagsResult) ListDerivedTagsResponse {
	return ListDerivedTagsResponse{
		Items: NewDerivedTagDtos(src.Data.Items),
		Page:  src.Data.Page,
		Size:  src.Data.Size,
		Total: src.Data.Total,
	}
}
