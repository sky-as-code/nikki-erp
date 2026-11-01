package fieldmetadata

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

type FieldMetadataRepository interface {
	dynamicresource.CrudRepository
}

type FieldMetadataDomainService interface {
	dynamicresource.CrudDomainService
}

type FieldMetadataApplicationService interface {
	dynamicresource.CrudApplicationService
}

type (
	CreateFieldMetadataCommand = dynamicresource.CreateCommand
	UpdateFieldMetadataCommand = dynamicresource.UpdateCommand
	DeleteFieldMetadataCommand = dynamicresource.DeleteCommand
	GetFieldMetadataByIdQuery  = dynamicresource.GetByIdQuery
	SearchFieldMetadataQuery   = dynamicresource.SearchQuery
	FieldMetadataExistsQuery   = dynamicresource.ExistsQuery
)

type (
	CreateFieldMetadataResult  = dynamicresource.CreateResult
	UpdateFieldMetadataResult  = dynamicresource.MutateResult
	DeleteFieldMetadataResult  = dynamicresource.MutateResult
	GetFieldMetadataByIdResult = dynamicresource.GetOneResult
	SearchFieldMetadataResult  = dynamicresource.SearchResult
	FieldMetadataExistsResult  = dynamicresource.ExistsResult
)
