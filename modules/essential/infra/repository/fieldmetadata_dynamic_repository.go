package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

// NewFieldMetadataRepository is handed the composable default by the onion.
func NewFieldMetadataRepository(base dynamicresource.CrudRepository) itFieldMeta.FieldMetadataRepository {
	return &FieldMetadataRepositoryImpl{CrudRepository: base}
}

type FieldMetadataRepositoryImpl struct {
	dynamicresource.CrudRepository
}
