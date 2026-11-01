package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itModelMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

// NewModelMetadataDomainService is handed the composable default by the onion: plain CRUD.
func NewModelMetadataDomainService(base dynamicresource.CrudDomainService) itModelMeta.ModelMetadataDomainService {
	return &ModelMetadataDomainServiceImpl{CrudDomainService: base}
}

type ModelMetadataDomainServiceImpl struct {
	dynamicresource.CrudDomainService
}
