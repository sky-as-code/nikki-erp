package fieldmetadata

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type FieldMetadataRepository interface {
	composable.CrudRepository
}

type FieldMetadataDomainService interface {
	composable.CrudDomainService
}

type FieldMetadataApplicationService interface {
	composable.CrudApplicationService
}

type (
	CreateFieldMetadataCommand = composable.CreateCommand
	UpdateFieldMetadataCommand = composable.UpdateCommand
	DeleteFieldMetadataCommand = composable.DeleteCommand
	GetFieldMetadataByIdQuery  = composable.GetByIdQuery
	SearchFieldMetadataQuery   = composable.SearchQuery
	FieldMetadataExistsQuery   = composable.ExistsQuery
)

type (
	CreateFieldMetadataResult  = composable.CreateResult
	UpdateFieldMetadataResult  = composable.MutateResult
	DeleteFieldMetadataResult  = composable.MutateResult
	GetFieldMetadataByIdResult = composable.GetOneResult
	SearchFieldMetadataResult  = composable.SearchResult
	FieldMetadataExistsResult  = composable.ExistsResult
)
