package accounting

import (
	"go.bryk.io/pkg/errors"

	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerOrgSettings declares the tax policy an organization sets for its own trading: an
// org-level schema only, since the default rounding policy and tax-inclusive pricing are business
// unit decisions. Registration is idempotent, keyed by (module, level).
func registerOrgSettings(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.AccountingModuleName,
		Level:     c.LevelOrg,
		Schema:    accsettings.OrgSettingsSchemaBuilder().Build(),
	})
	if err != nil {
		return errors.Wrap(err, "registerOrgSettings")
	}
	// A rejected registration is a defect in this module's own declaration, so it fails the boot.
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "registerOrgSettings")
	}
	return nil
}
