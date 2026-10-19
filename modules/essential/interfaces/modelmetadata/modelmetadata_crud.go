package modelmetadata

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type ModelMetadataRepository interface {
	composable.CrudRepository
}

type ModelMetadataDomainService interface {
	composable.CrudDomainService
}

type ModelMetadataApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateModelMetadataCommand = composable.CreateCommand
	UpdateModelMetadataCommand = composable.UpdateCommand
	DeleteModelMetadataCommand = composable.DeleteCommand
	GetModelMetadataByIdQuery  = composable.GetByIdQuery
	SearchModelMetadataQuery   = composable.SearchQuery
	ModelMetadataExistsQuery   = composable.ExistsQuery
)

type (
	CreateModelMetadataResult  = composable.CreateResult
	UpdateModelMetadataResult  = composable.MutateResult
	DeleteModelMetadataResult  = composable.MutateResult
	GetModelMetadataByIdResult = composable.GetOneResult
	SearchModelMetadataResult  = composable.SearchResult
	ModelMetadataExistsResult  = composable.ExistsResult
)
