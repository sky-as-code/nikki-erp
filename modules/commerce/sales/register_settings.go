package sales

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerOrgSettings declares the commercial policy an organization sets for its own selling.
// Sales registers an org-level schema only. Registration is idempotent, keyed by (module, level),
// so it runs unconditionally on every boot.
func registerOrgSettings(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.SalesModuleName,
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
