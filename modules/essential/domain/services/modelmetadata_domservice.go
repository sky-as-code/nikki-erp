package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itModelMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/modelmetadata"
)

// NewModelMetadataDomainService is handed the composable default by the onion: plain CRUD.
func NewModelMetadataDomainService(base composable.CrudDomainService) itModelMeta.ModelMetadataDomainService {
	return &ModelMetadataDomainServiceImpl{CrudDomainService: base}
}

type ModelMetadataDomainServiceImpl struct {
	composable.CrudDomainService
}
