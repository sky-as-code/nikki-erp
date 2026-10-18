package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

// NewSettingsSchemaDomainService is handed the composable default by the onion. The resource is read-only
// over REST, so it carries no rule; the settings behaviour lives in SettingsDomainServiceImpl.
func NewSettingsSchemaDomainService(base composable.CrudDomainService) itSchema.SettingsSchemaDomainService {
	return &SettingsSchemaDomainServiceImpl{CrudDomainService: base}
}

type SettingsSchemaDomainServiceImpl struct {
	composable.CrudDomainService
}
