package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

// NewFieldMetadataRepository is handed the composable default by the onion.
func NewFieldMetadataRepository(base composable.CrudRepository) itFieldMeta.FieldMetadataRepository {
	return &FieldMetadataRepositoryImpl{CrudRepository: base}
}

type FieldMetadataRepositoryImpl struct {
	composable.CrudRepository
}
