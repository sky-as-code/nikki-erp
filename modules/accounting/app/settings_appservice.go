package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// NewSettingsApplicationServiceImpl validates Accounting's tenant- and org-level settings against the
// schemas it registers and the system currency rule, then persists them through core's settings port.
//
// Authorization is the writer's first step: update on settings_record for tenant and org, the
// actor's own row for user -- the same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(
	settingsSvc coreSettings.SettingsService, currencySvc itExt.CurrencyExtService,
) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelTenant: accsettings.TenantSettingsSchemaBuilder().Build(),
		coreSettings.LevelOrg:    accsettings.OrgSettingsSchemaBuilder().Build(),
	}
	return &SettingsApplicationServiceImpl{
		writer: coreSettings.NewModuleSettingsWriter(modconstants.AccountingModuleName, schemas, settingsSvc,
			newSystemCurrencyRule(currencySvc)),
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
