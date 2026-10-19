package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

// NewFieldMetadataApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewFieldMetadataApplicationService(base composable.CrudApplicationService) itFieldMeta.FieldMetadataApplicationService {
	return &FieldMetadataApplicationServiceImpl{CrudApplicationService: base}
}

type FieldMetadataApplicationServiceImpl struct {
	composable.CrudApplicationService
}
