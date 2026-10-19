package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

// NewEnumDomainService is handed the composable default by the onion: plain CRUD, the type exists
// so a rule has a home the day one is needed.
func NewEnumDomainService(base composable.CrudDomainService) itEnum.EnumDomainService {
	return &EnumDomainServiceImpl{CrudDomainService: base}
}

type EnumDomainServiceImpl struct {
	composable.CrudDomainService
}
