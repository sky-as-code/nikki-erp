// Package settings holds Sales' settings schemas. Unlike a model schema, a settings schema owns no
// table: its values are stored as settings_records rows by the settings module.
package settings

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The setting names Sales declares.
const (
	// OrgSettingMaxPaymentMethodsPerBill bounds how many payments may settle one bill.
	OrgSettingMaxPaymentMethodsPerBill = "max_payment_methods_per_bill"

	// OrgSettingReturnWindowDays is how long after a sale a return may be raised.
	OrgSettingReturnWindowDays = "return_window_days"

	// OrgSettingInvoiceIssueDelayMinutes is how long after a sale is settled the automatic issuance
	// job waits before raising its electronic invoice.
	//
	// A delay rather than issuing on payment, because the minutes after a sale are when it is most
	// likely to be corrected or reversed at the counter — and an issued VAT invoice cannot simply be
	// deleted afterwards. Waiting turns a would-be credit note into a sale that was never invoiced.
	OrgSettingInvoiceIssueDelayMinutes = "invoice_issue_delay_minutes"

	// OrgSettingAllowOverpayment permits taking more money than the bill is for.
	OrgSettingAllowOverpayment = "allow_overpayment"

	// OrgSettingAllowCashChange permits giving change back in cash.
	OrgSettingAllowCashChange = "allow_cash_change"

	// OrgSettingDraftOrderExpiryHours is how long a draft order survives untouched.
	OrgSettingDraftOrderExpiryHours = "draft_order_expiry_hours"

	// OrgSettingRoundingScale is the decimal place monetary totals are rounded to.
	OrgSettingRoundingScale = "rounding_scale"

	// OrgSettingDefaultTaxRate is DEPRECATED: accounting's tax master computes the rate now. The key
	// is kept only so stored values are not orphaned; nothing reads it.
	OrgSettingDefaultTaxRate = "default_tax_rate"

	// OrgSettingDefaultSalesTaxCode names the accounting tax applied to every sale line. One code per
	// organization is an interim until the per-product tax association exists. Empty means no tax and
	// the engine writes real zeros — distinct from an undetermined tax, which refuses the sale.
	OrgSettingDefaultSalesTaxCode = "default_sales_tax_code"

	// OrgSettingOutgoingOperationTypeId is the Inventory operation type a sale ships against, and
	// OrgSettingIncomingOperationTypeId the one a return is received against. Sales holds the id and
	// never names a warehouse or location; the operation type carries its own source and destination.
	// Empty means fulfilment is unconfigured and a fulfilment request is REFUSED rather than guessing
	// a type, which would move real goods out of the wrong place.
	OrgSettingOutgoingOperationTypeId = "outgoing_operation_type_id"
	OrgSettingIncomingOperationTypeId = "incoming_operation_type_id"
)

// The defaults, restated in Go as a fallback when the settings read fails or the value is absent.
// They deliberately duplicate the default_value entries in org_settings.json: the JSON governs the
// settings UI, these govern the code. DefaultsAgreeWithSchema asserts the two cannot drift.
const (
	DefaultMaxPaymentMethodsPerBill = int32(3)
	DefaultReturnWindowDays         = int32(30)
	DefaultAllowOverpayment         = false
	DefaultAllowCashChange          = true
	DefaultDraftOrderExpiryHours    = int32(24)

	// Two hours: long enough that same-visit corrections and reversals happen before a document
	// exists, short enough that a buyer who asked for an invoice is not left waiting for it.
	DefaultInvoiceIssueDelayMinutes = int32(120)
)

// DefaultRoundingScale is 0: VND has no minor unit, so a total is rounded to whole dong. It cannot
// be a JSON default_value — ModelField.Validate treats a numeric zero as ABSENT platform-wide, so a
// stored 0 reads back as unset and a min of 0 would reject nothing. The field is declared with min 1
// and zero lives here as the fallback when the setting is unset.
const DefaultRoundingScale = int32(0)

// DefaultTaxRate is DEPRECATED along with the setting it backs; tax is resolved by calling
// accounting. Kept so a caller still reading the deprecated setting gets a defined answer rather
// than a nil decimal.
var DefaultTaxRate = decimal.Zero

// DefaultSalesTaxCode is empty: an organization that has configured nothing is not taxed. Guessing
// a code would silently charge VAT nobody asked for, surfacing as a tax liability, not a bug report.
const DefaultSalesTaxCode = ""

// OrgSettingsSchemaName is the name Sales registers its org-level settings under. Not a table: the
// values live in settings_records.
const OrgSettingsSchemaName = "sales_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the commercial policy an organization sets for its own selling.
// The document declares no table_name or should_build_db and extends no base model: the basemodel
// mixins would inject tenant_id and audit columns onto a schema with no table to put them in.
// rounding_scale and the tax settings carry allow_override false because they change what money
// means — differing scales cannot be consolidated, and differing taxes disagree on what was owed.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
