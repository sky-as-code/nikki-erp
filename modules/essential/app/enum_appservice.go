package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

// NewEnumApplicationService is handed the composable default by the onion: the authorized CRUD.
func NewEnumApplicationService(base dynamicresource.CrudApplicationService) itEnum.EnumApplicationService {
	return &EnumApplicationServiceImpl{CrudApplicationService: base}
}

type EnumApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}
