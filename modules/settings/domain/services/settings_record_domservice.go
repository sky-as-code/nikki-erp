package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

// NewSettingsRecordDomainService is handed the composable default by the onion. The resource is read-only
// over REST, so it carries no rule; the settings behaviour lives in SettingsDomainServiceImpl.
func NewSettingsRecordDomainService(base composable.CrudDomainService) itRecord.SettingsRecordDomainService {
	return &SettingsRecordDomainServiceImpl{CrudDomainService: base}
}

type SettingsRecordDomainServiceImpl struct {
	composable.CrudDomainService
}
