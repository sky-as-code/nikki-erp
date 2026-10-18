package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

// NewEnumApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewEnumApplicationService(base composable.CrudApplicationService) itEnum.EnumApplicationService {
	return &EnumApplicationServiceImpl{CrudApplicationService: base}
}

type EnumApplicationServiceImpl struct {
	composable.CrudApplicationService
}
