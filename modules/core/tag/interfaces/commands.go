package tag

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

const tagTypePrefix = "tag_"

// Add prefix to mark this Enum as a Tag, because Enum implementation is reused as Tag.
func prependTagType(tagType *string) *string {
	if tagType == nil {
		return nil
	}
	tagTypeStr := tagTypePrefix + *tagType
	return &tagTypeStr
}

type CreateTagCommand struct {
	Label model.LangJson `json:"label"`
}

func (this CreateTagCommand) ToEnumCommand(tagType string) enum.CreateEnumCommand {
	return enum.CreateEnumCommand{
		EntityName: "tag",
		Label:      this.Label,
		Type:       *prependTagType(&tagType),
	}
}

type CreateTagResult = GetTagByIdResult

func NewCreateTagResult(result *enum.CreateEnumResult) *CreateTagResult {
	return NewGetTagByIdResult(result)
}

type UpdateTagCommand struct {
	Id    model.Id       `param:"id" json:"id"`
	Etag  model.Etag     `json:"etag,omitempty"`
	Label model.LangJson `json:"label"`
}

func (this UpdateTagCommand) ToEnumCommand() enum.UpdateEnumCommand {
	return enum.UpdateEnumCommand{
		Id:         this.Id,
		EntityName: "tag",
		Etag:       this.Etag,
		Label:      &this.Label,
	}
}

type UpdateTagResult = GetTagByIdResult

func NewUpdateTagResult(result *enum.UpdateEnumResult) *UpdateTagResult {
	return NewGetTagByIdResult(result)
}

type DeleteTagCommand struct {
	Id *model.Id `json:"id" param:"id"`
}

func (this DeleteTagCommand) ToEnumCommand(tagType string) enum.DeleteEnumCommand {
	return enum.DeleteEnumCommand{
		Id:         this.Id,
		EntityName: "tag",
		Type:       prependTagType(&tagType),
	}
}

type DeleteTagResultData struct {
	DeletedAt time.Time `json:"deletedAt,omitempty"`
}
type DeleteTagResult = model.OpResult[*DeleteTagResultData]

func NewDeleteTagResult(src *enum.DeleteEnumResult) *DeleteTagResult {
	var data *DeleteTagResultData
	if src.HasData {
		data = &DeleteTagResultData{
			DeletedAt: src.Data.DeletedAt,
		}
	}
	return &DeleteTagResult{
		ClientError: src.ClientError,
		Data:        data,
		HasData:     src.HasData,
	}
}

// TODO: I don't think of any use case where we need to check if a single tag exists.
// Uncomment these lines if someday we do need.
// type TagExistsQuery = enum.EnumExistsQuery
// type TagExistsResult = enum.EnumExistsResult

type TagExistsMultiQuery enum.EnumExistsMultiQuery

func (this TagExistsMultiQuery) ToEnumQuery() enum.EnumExistsMultiQuery {
	return enum.EnumExistsMultiQuery(this)
}

type TagExistsMultiResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"notExisting"`
}
type TagExistsMultiResult = model.OpResult[*TagExistsMultiResultData]

func NewTagExistsMultiResult(src *enum.EnumExistsMultiResult) *TagExistsMultiResult {
	if src == nil {
		return nil
	}
	var data *TagExistsMultiResultData
	if src.HasData {
		data = &TagExistsMultiResultData{
			Existing:    src.Data.Existing,
			NotExisting: src.Data.NotExisting,
		}
	}
	return &TagExistsMultiResult{
		ClientError: src.ClientError,
		Data:        data,
		HasData:     src.HasData,
	}
}

type GetTagByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetTagByIdQuery) ToEnumQuery() enum.GetEnumQuery {
	return enum.GetEnumQuery{
		Id:         &this.Id,
		EntityName: "tag",
	}
}

type GetTagByIdResult = model.OpResult[*Tag]

func NewGetTagByIdResult(src *enum.GetEnumResult) *GetTagByIdResult {
	return &GetTagByIdResult{
		ClientError: src.ClientError,
		Data:        EnumToTag(src.Data),
		HasData:     src.HasData,
	}
}

type ListTagsQuery struct {
	// Part of the Enum Label used for filtering
	PartialLabel *string             `json:"partialLabel" query:"partialLabel"`
	Page         *int                `json:"page" query:"page"`
	Size         *int                `json:"size" query:"size"`
	SortedByLang *model.LanguageCode `json:"sortedByLang" query:"sortedByLang"`
}

func (this ListTagsQuery) ToEnumQuery(tagType string) enum.ListEnumsQuery {
	return enum.ListEnumsQuery{
		EntityName:   "tag",
		PartialLabel: this.PartialLabel,
		Type:         prependTagType(&tagType),
		Page:         this.Page,
		Size:         this.Size,
		SortedByLang: this.SortedByLang,
	}
}

type ListTagsResultData = crud.PagedResult[Tag]
type ListTagsResult = model.OpResult[*ListTagsResultData]

func NewListTagsResultData(src *enum.ListEnumsResultData) *ListTagsResultData {
	if src == nil {
		return nil
	}
	return &ListTagsResultData{
		Items: EnumsToTags(src.Items),
		Page:  src.Page,
		Size:  src.Size,
		Total: src.Total,
	}
}

func NewListTagsResult(src *enum.ListEnumsResult) *ListTagsResult {
	return &ListTagsResult{
		ClientError: src.ClientError,
		Data:        NewListTagsResultData(src.Data),
		HasData:     src.HasData,
	}
}

type SearchTagsQuery enum.SearchEnumsQuery

func (this SearchTagsQuery) ToEnumQuery(tagType string) enum.SearchEnumsQuery {
	return enum.SearchEnumsQuery{
		EntityName: "tag",
		Graph:      this.Graph,
		Page:       this.Page,
		Size:       this.Size,
		TypePrefix: prependTagType(&tagType),
	}
}

type SearchTagsResult = ListTagsResult
