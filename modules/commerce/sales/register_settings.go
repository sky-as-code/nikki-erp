package sales

import (
	"go.bryk.io/pkg/errors"

	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerOrgSettings declares the commercial policy an organization sets for its own selling.
// Sales registers an org-level schema only. Registration is idempotent, keyed by (module, level),
// so it runs unconditionally on every boot.
func registerOrgSettings(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.SettingsModuleKey,
		Level:     c.LevelOrg,
		Schema:    salessettings.OrgSettingsSchemaBuilder().Build(),
	})
	if err != nil {
		return errors.Wrap(err, "registerOrgSettings")
	}
	// A rejected registration is a defect in this module's own declaration, not something a user
	// can correct, so it fails the boot.
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "registerOrgSettings")
	}
	return nil
}
