// Package sequenceusage helps a Commerce part answer Essential's "is this sequence definition still
// in use" (usagecheck.ResourceSequenceDefinition), asked before a definition is archived.
//
// Every Commerce definition id lives in an org setting of the part that registered it (the
// transfer and scrap definitions of Inventory, the order and agreement ones of Purchase, Sales'
// document series), so NewSettingsChecker below is the one checker a part hands to Commerce's
// registry. It reads only the part's own settings, and reports a failed read as an error, never as
// "not used": an unanswered check must refuse the archive.
package sequenceusage

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// OrgSettingsReader reads the values one organization stores under a module key.
type OrgSettingsReader = coreSettings.OrgSettingsReader

// ReaderResolver returns the reader at check time: checkers are registered during Init, before the
// container can build the settings services.
type ReaderResolver func() (OrgSettingsReader, error)

// NewSettingsChecker answers "used" when one of the named org settings of moduleKey holds the
// definition id. Settings-held definitions are org-scoped, so a definition without an org (a
// tenant-wide one) is never held here and answers "not used".
func NewSettingsChecker(resolve ReaderResolver, moduleKey string, settingNames ...string) usagecheck.ResourceUsageChecker {
	return usagecheck.CheckerFunc(func(ctx corectx.Context, ref usagecheck.ResourceRef) (bool, string, error) {
		id := ref.Identifier[usagecheck.IdentifierKeyId]
		orgId := ref.Identifier[usagecheck.IdentifierKeyOrgId]
		if id == "" || orgId == "" || len(settingNames) == 0 {
			return false, "", nil
		}
		reader, err := resolve()
		if err != nil {
			return false, "", errors.Wrap(err, "resolve the settings reader")
		}
		values, err := reader.GetOrgSettingValues(ctx, model.Id(orgId), moduleKey)
		if err != nil {
			return false, "", errors.Wrapf(err, "read %s settings to check sequence %s", moduleKey, id)
		}
		for _, name := range settingNames {
			if held, ok := values[name].(string); ok && held == id {
				return true, moduleKey + "." + name, nil
			}
		}
		return false, "", nil
	})
}
