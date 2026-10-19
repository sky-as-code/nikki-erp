package tag

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type TagRepository interface {
	composable.CrudRepository
}

type TagDomainService interface {
	composable.CrudDomainService
}

type TagApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateTagCommand = composable.CreateCommand
	UpdateTagCommand = composable.UpdateCommand
	DeleteTagCommand = composable.DeleteCommand
	GetTagByIdQuery  = composable.GetByIdQuery
	SearchTagsQuery  = composable.SearchQuery
	TagExistsQuery   = composable.ExistsQuery
)

type (
	CreateTagResult  = composable.CreateResult
	UpdateTagResult  = composable.MutateResult
	DeleteTagResult  = composable.MutateResult
	GetTagByIdResult = composable.GetOneResult
	SearchTagsResult = composable.SearchResult
	TagExistsResult  = composable.ExistsResult
)
