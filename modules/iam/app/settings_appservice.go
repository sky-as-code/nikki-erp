package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	modconstants "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	iamsettings "github.com/sky-as-code/nikki-erp/modules/iam/domain/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/settings"
)

// NewSettingsApplicationServiceImpl validates Iam's tenant-level settings against the schema it
// registers, then persists them through core's settings port. Iam has no rule beyond its schema
// yet; one goes in as a coreSettings.SettingsRule argument.
//
// Authorization is the writer's first step: update on settings_record for tenant and org, the
// actor's own row for user -- the same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(settingsSvc coreSettings.SettingsService) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelTenant: iamsettings.TenantSettingsSchemaBuilder().Build(),
	}
	return &SettingsApplicationServiceImpl{
		writer: coreSettings.NewModuleSettingsWriter(modconstants.IamModuleName, schemas, settingsSvc),
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
