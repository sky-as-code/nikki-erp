package tag

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type TagRepository interface {
	dynamicresource.CrudRepository
}

type TagDomainService interface {
	dynamicresource.CrudDomainService
}

type TagApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateTagCommand = dynamicresource.CreateCommand
	UpdateTagCommand = dynamicresource.UpdateCommand
	DeleteTagCommand = dynamicresource.DeleteCommand
	GetTagByIdQuery  = dynamicresource.GetByIdQuery
	SearchTagsQuery  = dynamicresource.SearchQuery
	TagExistsQuery   = dynamicresource.ExistsQuery
)

type (
	CreateTagResult  = dynamicresource.CreateResult
	UpdateTagResult  = dynamicresource.MutateResult
	DeleteTagResult  = dynamicresource.MutateResult
	GetTagByIdResult = dynamicresource.GetOneResult
	SearchTagsResult = dynamicresource.SearchResult
	TagExistsResult  = dynamicresource.ExistsResult
)
