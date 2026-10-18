package app

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

// NewSettingsRecordApplicationService is handed the composable default by the onion: the authorized reads.
func NewSettingsRecordApplicationService(base composable.CrudApplicationService) itRecord.SettingsRecordApplicationService {
	return &SettingsRecordApplicationServiceImpl{CrudApplicationService: base}
}

type SettingsRecordApplicationServiceImpl struct {
	composable.CrudApplicationService
}
