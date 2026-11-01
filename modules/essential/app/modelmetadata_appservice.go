package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itModelMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

// NewModelMetadataApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewModelMetadataApplicationService(
	base dynamicresource.CrudApplicationService,
) itModelMeta.ModelMetadataApplicationService {
	return &ModelMetadataApplicationServiceImpl{CrudApplicationService: base}
}

type ModelMetadataApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
