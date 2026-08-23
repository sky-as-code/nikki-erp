package iam

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	iamsettings "github.com/sky-as-code/nikki-erp/modules/iam/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/external"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
)

// registerTenantSettings declares the authentication policy a tenant sets for everyone in it.
//
// iam registers a tenant-level schema only: a session timeout and an MFA requirement are decisions
// the tenant makes about everybody, not choices an organization or a person makes for themselves.
//
// Registration is idempotent, so it runs unconditionally on every boot.
func registerTenantSettings(ctx corectx.Context, settingsSvc itExt.SettingsRegistrationExtService) error {
	result, err := settingsSvc.RegisterSchema(ctx, itExt.RegisterSchemaCommand{
		ModuleKey: modconstants.IamModuleName,
		Level:     c.LevelTenant,
		Schema:    iamsettings.TenantSettingsSchemaBuilder().Build(),
	})
	if err != nil {
		return errors.Wrap(err, "registerTenantSettings")
	}
	// A rejected registration is a defect in this module's own declaration, not something a user
	// can correct, so it fails the boot rather than being reported to a caller who is not there.
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "registerTenantSettings")
	}
	return nil
}
