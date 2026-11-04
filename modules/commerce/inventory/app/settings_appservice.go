package app

import (
	"slices"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// NewSettingsApplicationServiceImpl validates Inventory's org-level settings against the schema it
// registers and checks every unit setting against Essential's units, then persists them through
// core's settings port.
//
// Authorization is the writer's first step: update on settings_record for tenant and org -- the
// same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(
	settingsSvc coreSettings.SettingsService, units *services.OrgUnitResolver,
) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelOrg: invsettings.OrgSettingsSchemaBuilder().Build(),
	}
	writer := coreSettings.NewModuleSettingsWriter(
		invconstants.InventoryModuleName, schemas, settingsSvc, serverOwnedSettingRule, unitSettingRule(units),
	)
	return &SettingsApplicationServiceImpl{writer: writer}
}

type SettingsApplicationServiceImpl struct {
	writer *coreSettings.ModuleSettingsWriter
}

func (this *SettingsApplicationServiceImpl) UpdateSettings(
	ctx corectx.Context, cmd coreSettings.UpdateSettingsCommand,
) (*coreSettings.SetSettingsResult, error) {
	return this.writer.Update(ctx, cmd)
}

// serverOwnedSettingRule refuses a client write to a setting only Inventory writes: the scrap
// sequence id. Re-pointing it would number scraps from another definition and hide the real one
// from the usage check that keeps it from being archived.
func serverOwnedSettingRule(_ corectx.Context, _ string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
	vErrs := ft.ClientErrors{}
	for _, item := range items {
		if slices.Contains(invsettings.ServerOwnedOrgSettings, item.Name) {
			vErrs.Append(*ft.NewValidationError(item.Name, ft.ErrorKey("settings.server_owned", "inventory"),
				"{{field}} is set by the system", map[string]any{"field": item.Name}))
		}
	}
	return vErrs, nil
}

// unitSettingRule refuses a unit that does not exist, is archived, or is not of the kind the
// setting measures. An empty value clears the setting, which then reads as the organization's
// default unit.
func unitSettingRule(units *services.OrgUnitResolver) coreSettings.SettingsRule {
	return func(ctx corectx.Context, _ string, items []coreSettings.SetSettingItem) (ft.ClientErrors, error) {
		vErrs := ft.ClientErrors{}
		for _, item := range items {
			uomId, isString := item.Value.(string)
			if _, isUnit := invsettings.DefaultUomSymbolOf[item.Name]; !isUnit || !isString || uomId == "" {
				continue
			}
			cErr, err := units.AssertUnitSetting(ctx, item.Name, uomId)
			if err != nil {
				return nil, err
			}
			if cErr != nil {
				vErrs.Append(*cErr)
			}
		}
		return vErrs, nil
	}
}
