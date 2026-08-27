// Package settings holds Sales' settings schemas.
//
// These are separate from domain/models because they are a different kind of thing: a model schema
// maps to a database table this module owns, while a settings schema is metadata only — it owns no
// table, and its values are stored as settings_records rows by the settings module.
package settings

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The setting names Sales declares. Every one of them is a policy the business requirement refers
// to but gives no home; declaring them here is what keeps them from being scattered as constants
// that nobody can change without a deploy.
const (
	// OrgSettingMaxPaymentMethodsPerBill bounds how many payments may settle one bill (BR §40).
	OrgSettingMaxPaymentMethodsPerBill = "max_payment_methods_per_bill"

	// OrgSettingReturnWindowDays is how long after a sale a return may be raised (BR §78).
	OrgSettingReturnWindowDays = "return_window_days"

	// OrgSettingAllowOverpayment permits taking more money than the bill is for (BR §42).
	OrgSettingAllowOverpayment = "allow_overpayment"

	// OrgSettingAllowCashChange permits giving change back in cash (BR §42).
	OrgSettingAllowCashChange = "allow_cash_change"

	// OrgSettingDraftOrderExpiryHours is how long a draft order survives untouched (BR §87.3).
	OrgSettingDraftOrderExpiryHours = "draft_order_expiry_hours"

	// OrgSettingRoundingScale is the decimal place monetary totals are rounded to (D-01).
	OrgSettingRoundingScale = "rounding_scale"

	// OrgSettingDefaultTaxRate is DEPRECATED and no longer feeds pricing.
	//
	// It was the D-38 seam: with no tax master in the codebase, an organization that had to charge a
	// flat rate set one here. A real tax master now exists in accounting, is seeded with the
	// Vietnamese VAT rates, and computes the rate itself. The key is kept so stored values are not
	// orphaned and an administrator can see what was configured; nothing reads it.
	OrgSettingDefaultTaxRate = "default_tax_rate"

	// OrgSettingDefaultSalesTaxCode names the accounting tax applied to every sale line.
	//
	// A single code for the whole organization is an interim. Doc 3 of the tax requirement specifies
	// a per-product association — product_template_sales_taxes, effective_sales_tax_ids — which is
	// not built in essential or inventory yet, so nothing can yet say that this product is 8% and
	// that one 5%. Until it exists, this is what Sales has to go on.
	//
	// Empty means no tax: the pricing engine writes real zeros. That is different from an
	// undetermined tax, which refuses the sale rather than recording a zero.
	OrgSettingDefaultSalesTaxCode = "default_sales_tax_code"

	// OrgSettingOutgoingOperationTypeId is the Inventory operation type a sale ships against
	// (SALES-049), and OrgSettingIncomingOperationTypeId the one a return is received against.
	//
	// Sales holds the ID and nothing else. It never names a warehouse or a location: the operation
	// type carries its own default source and destination, so choosing the type is the whole of the
	// decision Sales gets to make, and where the goods actually sit stays Inventory's (BR 3.2).
	//
	// **Overridable, unlike the tax settings.** Two organizations shipping from different warehouses
	// is ordinary — it is the normal reason to run more than one org — whereas two taxing the same
	// product differently would produce fiscal documents that disagree about what was owed.
	//
	// Empty means fulfilment is not configured. A fulfilment request is then REFUSED with a reason
	// naming the gap, rather than guessing a type: picking the wrong one moves real goods out of the
	// wrong place, and that is not a mistake worth a default.
	OrgSettingOutgoingOperationTypeId = "outgoing_operation_type_id"
	OrgSettingIncomingOperationTypeId = "incoming_operation_type_id"
)

// The defaults, restated in Go so that a caller reading a setting has something to fall back on
// when the read fails or the value is absent.
//
// They duplicate the default_value entries in org_settings.json deliberately. A settings read is
// a call into another module that can fail, and pricing an order must not fail with it; the JSON
// default governs what an administrator sees in the settings UI, and these govern what the code
// does when it has no answer. DefaultsAgreeWithSchema in the test file asserts the two match, so
// the duplication cannot drift.
const (
	DefaultMaxPaymentMethodsPerBill = int32(3)
	DefaultReturnWindowDays         = int32(30)
	DefaultAllowOverpayment         = false
	DefaultAllowCashChange          = true
	DefaultDraftOrderExpiryHours    = int32(24)
)

// DefaultRoundingScale is 0: VND has no minor unit, so a total is rounded to whole dong.
//
// It is NOT declared as a default_value in the JSON, and cannot usefully be. ModelField.Validate
// treats a numeric zero as an ABSENT value platform-wide (isNilOrEmpty returns true for it), so a
// stored 0 reads back as "no value chosen" and a `min` of 0 would never reject anything anyway.
// The field is therefore declared with min 1 — the range that is actually enforceable — and zero
// lives here as the fallback a caller uses when the setting is unset.
//
// The practical effect is the right one: an organization that has said nothing gets whole-dong
// rounding, and one that trades in a currency with minor units says so by setting 1 or 2. What is
// lost is the ability to distinguish "explicitly chose 0" from "never answered", and for this
// setting those two mean the same thing.
const DefaultRoundingScale = int32(0)

// DefaultTaxRate is zero and is DEPRECATED along with the setting it backs.
//
// D-38 recorded that no tax master existed anywhere in the codebase, which was true when the survey
// was taken and is not true now: accounting owns one, seeded with the Vietnamese VAT rates. Tax is
// resolved by calling that module, not by reading a rate from here.
//
// The value is kept so that a caller still reading the deprecated setting gets a defined answer
// rather than a nil decimal.
var DefaultTaxRate = decimal.Zero

// DefaultSalesTaxCode is empty: an organization that has configured nothing is not taxed.
//
// Empty is the honest default. Guessing a code — VN_VAT_10, say — would silently charge 10% VAT to
// a deployment that never asked for it and may not be Vietnamese, and the error would surface as a
// tax liability rather than as a broken screen.
const DefaultSalesTaxCode = ""

// OrgSettingsSchemaName is the name Sales registers its org-level settings under. It is not a
// table: the schema describes values that settings_records stores, and only the settings module
// owns tables.
const OrgSettingsSchemaName = "sales_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the commercial policy an organization sets for its own selling.
//
// The document declares no table_name and no should_build_db, and extends no base model: a settings
// schema is metadata only, and the core.basemodel mixins would inject tenant_id and audit columns
// onto a schema with no table to put them in.
//
// **Org rather than tenant.** These are decisions a business unit makes about how it sells, not
// tenant-wide security policy: a vending estate and a staffed store chain within one business
// genuinely differ on whether cash change is possible, and a return window is a commercial promise
// each makes to its own customers. The two exceptions carry allow_override false and are explained
// below.
//
// **Why rounding_scale and the tax settings cannot be overridden.** Both change what money means
// rather than what is permitted. Two organizations rounding to different scales would produce
// totals that cannot be added together in a consolidated report; two applying different taxes to
// the same product would produce fiscal documents that disagree about what was owed. The other
// five are genuine per-organization policy and are overridable.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
