package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// OrgDefaultsLookupExtService resolves what a new organization's sales defaults name by code,
// since the ids differ per tenant and per organization.
type OrgDefaultsLookupExtService interface {
	CurrencyIdByCode(ctx corectx.Context, code string) (model.Id, error)
	// OperationTypeIdByCode finds an Inventory operation type of the organization. Inventory sets
	// its defaults up first, being a dependency of Sales.
	OperationTypeIdByCode(ctx corectx.Context, orgId model.Id, code string) (model.Id, error)
}

// OrgSettingsSeedExtService writes Sales' own settings for a new organization, from the org
// lifecycle.
type OrgSettingsSeedExtService interface {
	SeedOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

type SetSettingsCommand = itSettings.SetSettingsCommand
type SetSettingItem = itSettings.SetSettingItem
type SetSettingsResult = itSettings.SetSettingsResult
