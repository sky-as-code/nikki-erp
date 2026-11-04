package paymentinvoice

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

// registerOrgSettings declares the org settings holding the invoice sequence id, under
// the "paymentinvoice" module key. Registration is idempotent, keyed by (module, level), so it runs
// unconditionally on every boot.
func registerOrgSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.PaymentInvoiceModuleName,
		Level:     coreSettings.LevelOrg,
		Schema:    pinvsettings.OrgSettingsSchemaBuilder().Build(),
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
