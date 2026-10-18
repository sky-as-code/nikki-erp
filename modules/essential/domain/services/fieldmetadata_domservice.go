package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itFieldMeta "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/fieldmetadata"
)

// NewFieldMetadataDomainService is handed the composable default by the onion: plain CRUD.
func NewFieldMetadataDomainService(base composable.CrudDomainService) itFieldMeta.FieldMetadataDomainService {
	return &FieldMetadataDomainServiceImpl{CrudDomainService: base}
}

type FieldMetadataDomainServiceImpl struct {
	composable.CrudDomainService
}
