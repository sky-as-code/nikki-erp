package app

import (
	"slices"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itSettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// NewSettingsApplicationServiceImpl validates Sales's org-level settings against the schema it
// registers, then persists them through core's settings port. Its one rule beyond the schema is
// SequenceSettingsRule.
//
// Authorization is the writer's first step: update on settings_record for tenant and org, the
// actor's own row for user -- the same check the Settings module's application services run.
func NewSettingsApplicationServiceImpl(settingsSvc coreSettings.SettingsService) itSettings.SettingsAppService {
	schemas := map[string]*dmodel.ModelSchema{
		coreSettings.LevelOrg: salessettings.OrgSettingsSchemaBuilder().Build(),
	}
	return &SettingsApplicationServiceImpl{
		writer: coreSettings.NewModuleSettingsWriter(
			modconstants.SettingsModuleKey, schemas, settingsSvc, SequenceSettingsRule),
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

// SequenceSettingsRule refuses a client write to a setting holding a sequence definition id. The
// org lifecycle registers the definitions and records their ids; repointing one would restart a
// document series or share it with another, and the old numbers would collide with the new.
func SequenceSettingsRule(
	_ corectx.Context, _ string, items []coreSettings.SetSettingItem,
) (ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()
	for _, item := range items {
		if slices.Contains(salessettings.SequenceIdSettings(), item.Name) {
			vErrs.Append(*ft.NewBusinessViolation(item.Name, ft.ErrorKey("settings.sequence_managed", "sales"),
				"this setting is managed by the system and cannot be changed"))
		}
	}
	return *vErrs, nil
}
