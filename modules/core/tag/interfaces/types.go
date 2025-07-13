package tag

import (
	"context"

	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type TagServiceFactory func(tagType enum.EnumType) (TagService, error)

type TagService interface {
	CreateTag(ctx context.Context, cmd CreateTagCommand) (*CreateTagResult, error)
	DeleteTag(ctx context.Context, cmd DeleteTagCommand) (*DeleteTagResult, error)
	TagExistsMulti(ctx context.Context, cmd TagExistsMultiQuery) (*TagExistsMultiResult, error)
	GetTagById(ctx context.Context, query GetTagByIdQuery) (result *GetTagByIdResult, err error)
	ListTags(ctx context.Context, query ListTagsQuery) (result *ListTagsResult, err error)
	UpdateTag(ctx context.Context, cmd UpdateTagCommand) (*UpdateTagResult, error)
}
