package tag

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type TagService interface {
	CreateTag(ctx corectx.Context, cmd CreateTagCommand) (*CreateTagResult, error)
	UpdateTag(ctx corectx.Context, cmd UpdateTagCommand) (*UpdateTagResult, error)
	DeleteTag(ctx corectx.Context, cmd DeleteTagCommand) (*DeleteTagResult, error)
	GetTag(ctx corectx.Context, query GetTagQuery) (*GetTagResult, error)
	SearchTags(ctx corectx.Context, query SearchTagsQuery) (*SearchTagsResult, error)
	TagExists(ctx corectx.Context, query TagExistsQuery) (*TagExistsResult, error)
}
