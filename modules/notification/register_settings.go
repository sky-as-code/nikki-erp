package notification

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerSettings declares what an organization may configure about its own notifications.
//
// Only an org level is registered. There is nothing here a single person should set for themselves
// — whether the web channel is on, and how the stream behaves, are properties of the deployment's
// organization, not preferences. Per-user notification preferences are explicitly future work.
func registerSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	return registerSchema(ctx, settingsSvc, c.LevelOrg, orgSettingsSchema())
}

func registerSchema(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
	level string, schema *dmodel.ModelSchema,
) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.NotificationModuleName,
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
