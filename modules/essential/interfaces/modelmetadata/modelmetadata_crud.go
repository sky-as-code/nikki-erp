package modelmetadata

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type ModelMetadataRepository interface {
	dynamicresource.CrudRepository
}

type ModelMetadataDomainService interface {
	dynamicresource.CrudDomainService
}

type ModelMetadataApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateModelMetadataCommand = dynamicresource.CreateCommand
	UpdateModelMetadataCommand = dynamicresource.UpdateCommand
	DeleteModelMetadataCommand = dynamicresource.DeleteCommand
	GetModelMetadataByIdQuery  = dynamicresource.GetByIdQuery
	SearchModelMetadataQuery   = dynamicresource.SearchQuery
	ModelMetadataExistsQuery   = dynamicresource.ExistsQuery
)

type (
	CreateModelMetadataResult  = dynamicresource.CreateResult
	UpdateModelMetadataResult  = dynamicresource.MutateResult
	DeleteModelMetadataResult  = dynamicresource.MutateResult
	GetModelMetadataByIdResult = dynamicresource.GetOneResult
	SearchModelMetadataResult  = dynamicresource.SearchResult
	ModelMetadataExistsResult  = dynamicresource.ExistsResult
)
