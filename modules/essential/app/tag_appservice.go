package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itTag "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

// NewTagApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewTagApplicationService(base composable.CrudApplicationService) itTag.TagApplicationService {
	return &TagApplicationServiceImpl{CrudApplicationService: base}
}

type TagApplicationServiceImpl struct {
	composable.CrudApplicationService
}
