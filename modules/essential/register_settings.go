package essential

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	settings "github.com/sky-as-code/nikki-erp/modules/essential/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerSettings declares what Essential can be configured with, at both levels it owns.
//
// Essential registers no tenant-level schema: nothing it owns is a tenant-wide policy. A tenant can
// still enforce any of these values downward, but that comes from the settings module's override
// rule rather than from a tenant-level declaration here.
//
// Registration is idempotent, so it runs unconditionally on every boot.
func registerSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	return stdErr.Join(
		registerSchema(ctx, settingsSvc, c.LevelUser, settings.UserSettingsSchemaBuilder().Build()),
		registerSchema(ctx, settingsSvc, c.LevelOrg, settings.OrgSettingsSchemaBuilder().Build()),
	)
}

func registerSchema(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
	level string, schema *dmodel.ModelSchema,
) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.EssentialModuleName,
		Level:     level,
		Schema:    schema,
	})
	if err != nil {
		return errors.Wrapf(err, "registerSettings: %s level", level)
	}
	// A rejected registration is a defect in this module's own declaration, not something a user
	// can correct, so it fails the boot rather than being reported to a caller who is not there.
	if result.ClientErrors.Count() > 0 {
		return errors.Wrapf(result.ClientErrors.ToError(), "registerSettings: %s level", level)
	}
	return nil
}
