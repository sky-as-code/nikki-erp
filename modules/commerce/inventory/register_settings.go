package inventory

import (
	"go.bryk.io/pkg/errors"

	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// registerOrgSettings declares the units an organization measures its products in, under the
// "inventory" module key. Registration is idempotent, keyed by (module, level), so it runs
// unconditionally on every boot.
func registerOrgSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: invconstants.InventoryModuleName,
		Level:     coreSettings.LevelOrg,
		Schema:    invsettings.OrgSettingsSchemaBuilder().Build(),
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
