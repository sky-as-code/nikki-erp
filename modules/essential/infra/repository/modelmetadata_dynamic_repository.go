package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itModelMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

// NewModelMetadataRepository is handed the composable default by the onion.
func NewModelMetadataRepository(base composable.CrudRepository) itModelMeta.ModelMetadataRepository {
	return &ModelMetadataRepositoryImpl{CrudRepository: base}
}

type ModelMetadataRepositoryImpl struct {
	composable.CrudRepository
}
