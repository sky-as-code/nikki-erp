// Package settings declares what an organization may configure about its own tax handling.
//
// These are not domain models: a settings schema owns no table and its values live in the settings
// module's storage, which is why the JSON carries no table_name, no should_build_db and none of the
// core.basemodel mixins.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

const (
	OrgSettingsSchemaName = "accounting_org_settings"

	// OrgSettingOrgDefaultCurrency is the currency an organization keeps its books in: the one
	// every unqualified monetary amount it stores is denominated in.
	//
	// The value is an `essential_currency` id, not the three-letter code, because Essential
	// resolves a currency by id only and storing the code would mean no validation.
	//
	// This is the single place the answer exists; Product's base_sales_price and cost carry no
	// currency of their own. Empty means unconfigured, and a caller that needs the currency must
	// refuse rather than guess, since misreading an amount's currency fails silently.
	OrgSettingOrgDefaultCurrency = "org_default_currency"

	// OrgSettingDefaultRoundingPolicyCode names the rounding policy used when a calculation
	// request does not name one. Empty means there is no default, and a request that needs a
	// policy and names none fails as unresolved rather than falling back to a hardcoded scale.
	OrgSettingDefaultRoundingPolicyCode = "default_rounding_policy_code"

	// OrgSettingDefaultPriceMode is the price mode a request inherits when it does not state one,
	// and what a tax whose price_inclusion_mode is "inherit" resolves against.
	OrgSettingDefaultPriceMode = "default_price_mode"

	// OrgSettingRequireOverrideReason lets an organization additionally refuse an empty reason on a
	// manual tax override, which already requires a reason.
	OrgSettingRequireOverrideReason = "require_override_reason"
)

//go:embed org_settings.json
var orgSettingsSchemaJson string

func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
