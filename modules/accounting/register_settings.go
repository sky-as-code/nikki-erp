package accounting

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	accsettings "github.com/sky-as-code/nikki-erp/modules/accounting/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/accounting/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerSettings declares the tenant's default system currency and the policy an organization
// sets for its own trading: the default rounding policy and tax-inclusive pricing are business
// unit decisions. Registration is idempotent, keyed by (module, level).
func registerSettings(
	ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService,
) error {
	schemas := map[string]*dmodel.ModelSchema{
		c.LevelTenant: accsettings.TenantSettingsSchemaBuilder().Build(),
		c.LevelOrg:    accsettings.OrgSettingsSchemaBuilder().Build(),
	}
	for _, level := range []string{c.LevelTenant, c.LevelOrg} {
		result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
			ModuleKey: modconstants.AccountingModuleName,
			Level:     level,
			Schema:    schemas[level],
		})
		if err != nil {
			return errors.Wrap(err, "registerSettings")
		}
		// A rejected registration is a defect in this module's own declaration, so it fails the boot.
		if result.ClientErrors.Count() > 0 {
			return errors.Wrap(result.ClientErrors.ToError(), "registerSettings")
		}
	}
	return nil
}
