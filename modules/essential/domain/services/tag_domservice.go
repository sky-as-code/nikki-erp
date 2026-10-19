package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itTag "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

// NewTagDomainService is handed the composable default by the onion: plain CRUD, the type exists
// so a rule has a home the day one is needed.
func NewTagDomainService(base composable.CrudDomainService) itTag.TagDomainService {
	return &TagDomainServiceImpl{CrudDomainService: base}
}

type TagDomainServiceImpl struct {
	composable.CrudDomainService
}
