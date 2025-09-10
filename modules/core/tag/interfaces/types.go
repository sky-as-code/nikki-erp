package tag

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type TagServiceFactory func(tagType enum.EnumType) (TagService, error)

type TagService interface {
	CreateTag(ctx crud.Context, cmd CreateTagCommand) (*CreateTagResult, error)
	DeleteTag(ctx crud.Context, cmd DeleteTagCommand) (*DeleteTagResult, error)
	TagExistsMulti(ctx crud.Context, cmd TagExistsMultiQuery) (*TagExistsMultiResult, error)
	GetTagById(ctx crud.Context, query GetTagByIdQuery) (result *GetTagByIdResult, err error)
	ListTags(ctx crud.Context, query ListTagsQuery) (result *ListTagsResult, err error)
	UpdateTag(ctx crud.Context, cmd UpdateTagCommand) (*UpdateTagResult, error)
}
