package services

import (
	"github.com/shopspring/decimal"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// SalesPolicy is one resolved snapshot of the settings that govern selling.
//
// Plain values rather than a live accessor, so a rule reads the policy once and uses the same
// numbers throughout: an accessor re-read mid-calculation could return a different rounding scale
// between two lines of the same order, producing totals that do not add up.
type SalesPolicy struct {
	MaxPaymentMethodsPerBill int32
	ReturnWindowDays         int32
	AllowOverpayment         bool
	AllowCashChange          bool
	DraftOrderExpiryHours    int32
	RoundingScale            int32

	// InvoiceIssueDelayMinutes is how long after settlement the automatic issuance job waits before
	// raising an electronic invoice. See domain/settings for why the wait exists.
	InvoiceIssueDelayMinutes int32

	// DefaultTaxRate is deprecated and no longer feeds pricing; see domain/settings. Still resolved
	// so an administrator reading the policy sees what is stored.
	DefaultTaxRate decimal.Decimal

	// DefaultSalesTaxCode names the accounting tax applied to every sale line. Empty means untaxed.
	DefaultSalesTaxCode string

	// The Inventory operation types a sale ships against and a return is received against. Empty
	// means fulfilment is not configured, and a fulfilment request is refused rather than guessed:
	// unlike every other field here there is no safe default, because an operation type decides
	// which warehouse the goods leave.
	OutgoingOperationTypeId string
	IncomingOperationTypeId string
}

// DefaultSalesPolicy is what applies when nothing has been configured, and what a failed settings
// read falls back to. Every field is the documented default from domain/settings.
func DefaultSalesPolicy() SalesPolicy {
	return SalesPolicy{
		MaxPaymentMethodsPerBill: salessettings.DefaultMaxPaymentMethodsPerBill,
		ReturnWindowDays:         salessettings.DefaultReturnWindowDays,
		AllowOverpayment:         salessettings.DefaultAllowOverpayment,
		AllowCashChange:          salessettings.DefaultAllowCashChange,
		DraftOrderExpiryHours:    salessettings.DefaultDraftOrderExpiryHours,
		RoundingScale:            salessettings.DefaultRoundingScale,
		InvoiceIssueDelayMinutes: salessettings.DefaultInvoiceIssueDelayMinutes,
		DefaultTaxRate:           salessettings.DefaultTaxRate,
	}
}

// ResolveSalesPolicy reads the settings that apply to the caller.
//
// It never fails: a settings read is a call into another module, and pricing an order must not stop
// because that module is slow, mid-restart or has no rows yet, so an unreadable setting falls back
// to its default and the sale proceeds. The trade is that a misconfigured settings module looks
// like a correctly configured default one, which is acceptable only because every default is a safe
// reading (overpayment off, narrowest rounding, no tax code). Falling back to no tax code makes the
// sale untaxed, which is different from a tax Accounting could not determine: that case refuses the
// sale in ResolveBasketTax and never reaches a default.
func ResolveSalesPolicy(
	ctx corectx.Context, settings itExt.EffectiveSettingsExtService,
) SalesPolicy {
	policy := DefaultSalesPolicy()
	if settings == nil {
		return policy
	}

	result, err := settings.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{modconstants.SalesModuleName},
	})
	if err != nil || result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return policy
	}
	values := result.Data.Values

	policy.MaxPaymentMethodsPerBill = int32Setting(values,
		salessettings.OrgSettingMaxPaymentMethodsPerBill, policy.MaxPaymentMethodsPerBill)
	policy.ReturnWindowDays = int32Setting(values,
		salessettings.OrgSettingReturnWindowDays, policy.ReturnWindowDays)
	policy.AllowOverpayment = boolSetting(values,
		salessettings.OrgSettingAllowOverpayment, policy.AllowOverpayment)
	policy.AllowCashChange = boolSetting(values,
		salessettings.OrgSettingAllowCashChange, policy.AllowCashChange)
	policy.DraftOrderExpiryHours = int32Setting(values,
		salessettings.OrgSettingDraftOrderExpiryHours, policy.DraftOrderExpiryHours)
	policy.RoundingScale = int32Setting(values,
		salessettings.OrgSettingRoundingScale, policy.RoundingScale)
	policy.InvoiceIssueDelayMinutes = int32Setting(values,
		salessettings.OrgSettingInvoiceIssueDelayMinutes, policy.InvoiceIssueDelayMinutes)
	policy.DefaultTaxRate = decimalSetting(values,
		salessettings.OrgSettingDefaultTaxRate, policy.DefaultTaxRate)
	policy.DefaultSalesTaxCode = stringSetting(values,
		salessettings.OrgSettingDefaultSalesTaxCode, policy.DefaultSalesTaxCode)
	policy.OutgoingOperationTypeId = stringSetting(values,
		salessettings.OrgSettingOutgoingOperationTypeId, policy.OutgoingOperationTypeId)
	policy.IncomingOperationTypeId = stringSetting(values,
		salessettings.OrgSettingIncomingOperationTypeId, policy.IncomingOperationTypeId)

	return policy
}

// settingKey builds the "{module_key}.{setting_name}" key GetEffectiveSettings flattens to.
func settingKey(name string) string {
	return modconstants.SalesModuleName + "." + name
}

// The readers below all take a fallback and never report failure, for the same reason
// ResolveSalesPolicy does not: refusing to sell is a worse response to a configuration defect than
// selling under the documented default. They also never bare type-assert, because a value that has
// been through a jsonb column arrives as whatever the JSON decoder chose (a whole number is a
// float64, not an int).

func int32Setting(values map[string]any, name string, fallback int32) int32 {
	value, ok := values[settingKey(name)]
	if !ok || value == nil {
		return fallback
	}
	switch typed := value.(type) {
	case int32:
		return typed
	case int64:
		return int32(typed)
	case int:
		return int32(typed)
	case float64:
		// The usual case: jsonb has no integer type, so every number comes back as a float.
		return int32(typed)
	case float32:
		return int32(typed)
	}
	return fallback
}

func boolSetting(values map[string]any, name string, fallback bool) bool {
	value, ok := values[settingKey(name)]
	if !ok || value == nil {
		return fallback
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return fallback
}

func stringSetting(values map[string]any, name string, fallback string) string {
	value, ok := values[settingKey(name)]
	if !ok || value == nil {
		return fallback
	}
	switch typed := value.(type) {
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
	}
	return fallback
}

func decimalSetting(
	values map[string]any, name string, fallback decimal.Decimal,
) decimal.Decimal {
	value, ok := values[settingKey(name)]
	if !ok || value == nil {
		return fallback
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed != nil {
			return *typed
		}
	case string:
		// A decimal crosses JSON as a string so it does not lose precision. Parsing it back is the
		// one place that could fail, and a malformed one falls back rather than propagating: a tax
		// rate that will not parse must not become NaN.
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	}
	return fallback
}
