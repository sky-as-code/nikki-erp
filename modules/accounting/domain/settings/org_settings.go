// Package settings declares what an organization may configure about its own tax handling.
//
// These are not domain models: a settings schema owns no table, and its values are rows in the
// settings module's own storage. That is why the JSON carries no table_name, no should_build_db
// and none of the core.basemodel mixins — those would inject audit columns and a tenant key for a
// table that does not exist.
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
	// The value is an `essential_currency` **id**, matching how Sales stores
	// outgoing_operation_type_id, and not the three-letter code. The code would read better in
	// the raw settings row, but Essential resolves a currency by id only — storing the code
	// would mean either a new lookup path there or no validation at all, and an unvalidated
	// currency is the failure BR-PRICE-CUR-004 names.
	//
	// It lives here rather than on each module that needs it because the alternative is worse:
	// a currency on Product, another on Inventory and a third on Sales would be three answers to
	// one question, free to disagree. Product's base_sales_price and cost carry no currency of
	// their own precisely so that this is the only place the answer exists.
	//
	// Empty means unconfigured. A caller that needs the currency must then refuse rather than
	// guess — reinterpreting an amount as the wrong currency is the failure BR-PRICE-CUR-004
	// names, and it is silent.
	OrgSettingOrgDefaultCurrency = "org_default_currency"

	// OrgSettingDefaultRoundingPolicyCode names the rounding policy used when a calculation
	// request does not name one. Empty means there is no default, and a request that needs a
	// policy and names none fails as unresolved rather than falling back to a hardcoded scale.
	OrgSettingDefaultRoundingPolicyCode = "default_rounding_policy_code"

	// OrgSettingDefaultPriceMode is the price mode a request inherits when it does not state one,
	// and what a tax whose price_inclusion_mode is "inherit" resolves against.
	OrgSettingDefaultPriceMode = "default_price_mode"

	// OrgSettingRequireOverrideReason lets an organization insist on a written reason for every
	// manual tax override. The requirement already mandates a reason (BR-TAX-ESS-024); this exists
	// so an organization can additionally refuse the empty string.
	OrgSettingRequireOverrideReason = "require_override_reason"
)

//go:embed org_settings.json
var orgSettingsSchemaJson string

func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
