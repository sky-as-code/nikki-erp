package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// NewSettingsApplicationServiceImpl validates Sales's org-level settings against the schema it
// registers, then persists them through core's settings port. Sales has no rule beyond its schema
// yet; one goes in as a coreSettings.SettingsRule argument.
//
// Authorization is the writer's first step: update on settings_record for tenant and org, the
// actor's own row for user -- the same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(settingsSvc coreSettings.SettingsService) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelOrg: salessettings.OrgSettingsSchemaBuilder().Build(),
	}
	return &SettingsApplicationServiceImpl{
		writer: coreSettings.NewModuleSettingsWriter(modconstants.SettingsModuleKey, schemas, settingsSvc),
	}
}

type SettingsApplicationServiceImpl struct {
	writer *coreSettings.ModuleSettingsWriter
}

func (this *SettingsApplicationServiceImpl) UpdateSettings(
	ctx corectx.Context, cmd coreSettings.UpdateSettingsCommand,
) (*coreSettings.SetSettingsResult, error) {
	return this.writer.Update(ctx, cmd)
}
