package purchase

import (
	"go.bryk.io/pkg/errors"

	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	pursettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
)

// registerOrgSettings declares the org settings holding Purchase's sequence ids, under the
// "purchase" module key. Registration is idempotent, keyed by (module, level), so it runs
// unconditionally on every boot.
func registerOrgSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: purconstants.PurchaseModuleName,
		Level:     coreSettings.LevelOrg,
		Schema:    pursettings.OrgSettingsSchemaBuilder().Build(),
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
