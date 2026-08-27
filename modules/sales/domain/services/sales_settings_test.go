package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	salessettings "github.com/sky-as-code/nikki-erp/modules/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// stubSettings answers whatever the test told it to, including by failing.
type stubSettings struct {
	values map[string]any
	err    error
	noData bool
	cErrs  bool
}

func (this *stubSettings) GetEffectiveSettings(
	ctx corectx.Context, query itExt.GetEffectiveSettingsQuery,
) (*itExt.GetEffectiveSettingsResult, error) {
	if this.err != nil {
		return nil, this.err
	}
	result := &itExt.GetEffectiveSettingsResult{
		HasData: !this.noData,
		Data:    itSettings.GetEffectiveSettingsResultData{Values: this.values},
	}
	if this.cErrs {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("settings", "settings.denied", "no"))
		result.ClientErrors = *vErrs
		result.HasData = false
	}
	return result, nil
}

func key(name string) string {
	return "sales." + name
}

// The whole contract of ResolveSalesPolicy: it never fails.
//
// A settings read is a call into another module, and pricing an order must not stop because that
// module is slow, mid-restart or has no rows yet. Each of these is a way the read can go wrong, and
// every one of them must produce the documented defaults rather than an error.
func TestResolveSalesPolicyNeverFails(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	want := DefaultSalesPolicy()

	cases := map[string]itExt.EffectiveSettingsExtService{
		"no port bound at all":  nil,
		"the call errors":       &stubSettings{err: errors.New("settings module is down")},
		"the call has no data":  &stubSettings{noData: true},
		"the call is refused":   &stubSettings{cErrs: true},
		"the values are absent": &stubSettings{values: map[string]any{}},
		"the values are nil":    &stubSettings{values: nil},
	}
	for name, settings := range cases {
		t.Run(name, func(t *testing.T) {
			got := ResolveSalesPolicy(ctx, settings)
			if got != want {
				t.Errorf("a failed settings read must yield the defaults, got %+v", got)
			}
		})
	}
}

// Every default must be the SAFE reading, which is what makes falling back on failure acceptable.
//
// This is the assumption the never-fails behaviour rests on: if a default were the permissive
// direction, a settings outage would silently grant something the organization had forbidden. It is
// asserted rather than left as prose because a later edit changing one default would break the
// reasoning without breaking anything else.
func TestEveryDefaultIsTheSafeReading(t *testing.T) {
	policy := DefaultSalesPolicy()

	if policy.AllowOverpayment {
		t.Error("overpayment must default to OFF: a settings outage must not start accepting " +
			"more money than a bill is for")
	}
	if policy.DefaultSalesTaxCode != "" {
		t.Error("the tax code must default to empty: an outage must not invent a tax charge, and " +
			"guessing a code would tax a deployment that never asked to be")
	}
	if !policy.DefaultTaxRate.IsZero() {
		// Deprecated and no longer read by pricing, but still resolved, so a caller reading it gets
		// a defined answer rather than a nil decimal.
		t.Error("the deprecated tax rate must still default to zero")
	}
	if policy.RoundingScale != 0 {
		t.Error("rounding must default to whole units, the narrowest reading for VND")
	}
	if policy.MaxPaymentMethodsPerBill < 1 {
		t.Error("a bill must always accept at least one payment method, or nothing can be paid")
	}
	if policy.ReturnWindowDays < 1 {
		t.Error("a zero return window would refuse every return during a settings outage")
	}
	// allow_cash_change defaults to TRUE, which is the one permissive-looking default. It is safe
	// in the direction that matters: the failure it prevents is a till that has taken cash and
	// cannot give change back, which strands a customer's money. Refusing to give change is the
	// harmful outcome here, not giving it.
	if !policy.AllowCashChange {
		t.Error("cash change must default to ON: a till that takes cash and cannot give change " +
			"back strands the customer's money")
	}
}

// Configured values must actually be read, and read through the "{module}.{name}" key.
//
// Numbers arrive as float64 because they have been through a jsonb column: encoding/json has no
// integer type. A reader that only accepted int32 would silently ignore every configured number and
// keep using its default, which no test of the defaults alone would catch.
func TestConfiguredValuesAreReadFromJsonShapes(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())

	policy := ResolveSalesPolicy(ctx, &stubSettings{values: map[string]any{
		key(salessettings.OrgSettingMaxPaymentMethodsPerBill): float64(5),
		key(salessettings.OrgSettingReturnWindowDays):         float64(14),
		key(salessettings.OrgSettingAllowOverpayment):         true,
		key(salessettings.OrgSettingAllowCashChange):          false,
		key(salessettings.OrgSettingDraftOrderExpiryHours):    float64(48),
		key(salessettings.OrgSettingRoundingScale):            float64(2),
		key(salessettings.OrgSettingDefaultTaxRate):           "0.1",
	}})

	if policy.MaxPaymentMethodsPerBill != 5 {
		t.Errorf("MaxPaymentMethodsPerBill = %d, want 5", policy.MaxPaymentMethodsPerBill)
	}
	if policy.ReturnWindowDays != 14 {
		t.Errorf("ReturnWindowDays = %d, want 14", policy.ReturnWindowDays)
	}
	if !policy.AllowOverpayment {
		t.Error("AllowOverpayment was configured true and must be read as true")
	}
	if policy.AllowCashChange {
		t.Error("AllowCashChange was configured false and must be read as false")
	}
	if policy.DraftOrderExpiryHours != 48 {
		t.Errorf("DraftOrderExpiryHours = %d, want 48", policy.DraftOrderExpiryHours)
	}
	if policy.RoundingScale != 2 {
		t.Errorf("RoundingScale = %d, want 2", policy.RoundingScale)
	}
	if !policy.DefaultTaxRate.Equal(decimal.RequireFromString("0.1")) {
		t.Errorf("DefaultTaxRate = %s, want 0.1", policy.DefaultTaxRate)
	}
}

// A decimal crosses JSON as a string so it does not lose precision on the way. Parsing it back is
// the one step that can fail, and a malformed rate must fall back rather than become NaN — a tax
// rate nobody can parse must not silently become a tax rate of something else.
func TestMalformedDecimalFallsBack(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())

	for _, bad := range []any{"not a number", "", []string{"0.1"}, true} {
		policy := ResolveSalesPolicy(ctx, &stubSettings{values: map[string]any{
			key(salessettings.OrgSettingDefaultTaxRate): bad,
		}})
		if !policy.DefaultTaxRate.Equal(salessettings.DefaultTaxRate) {
			t.Errorf("a %T of %v must fall back to the default rate, got %s",
				bad, bad, policy.DefaultTaxRate)
		}
	}
}

// A value of the wrong type is a configuration defect, and refusing to sell is a worse response to
// it than selling under the documented default. Each reader must ignore what it cannot use.
func TestWrongTypesFallBackPerSetting(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	want := DefaultSalesPolicy()

	policy := ResolveSalesPolicy(ctx, &stubSettings{values: map[string]any{
		key(salessettings.OrgSettingMaxPaymentMethodsPerBill): "five",
		key(salessettings.OrgSettingAllowCashChange):          "yes",
		key(salessettings.OrgSettingReturnWindowDays):         nil,
	}})

	if policy.MaxPaymentMethodsPerBill != want.MaxPaymentMethodsPerBill {
		t.Error("a string where an int was expected must fall back")
	}
	if policy.AllowCashChange != want.AllowCashChange {
		t.Error("a string where a bool was expected must fall back")
	}
	if policy.ReturnWindowDays != want.ReturnWindowDays {
		t.Error("an explicit nil must fall back")
	}
}

// Another module's settings must not be mistaken for this module's. The key carries the module
// prefix precisely so that two modules can both declare a setting called "rounding_scale" without
// one reading the other's.
func TestOtherModulesSettingsAreIgnored(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())
	want := DefaultSalesPolicy()

	policy := ResolveSalesPolicy(ctx, &stubSettings{values: map[string]any{
		"essential." + salessettings.OrgSettingRoundingScale: float64(3),
		salessettings.OrgSettingRoundingScale:                float64(4),
	}})

	if policy.RoundingScale != want.RoundingScale {
		t.Errorf("RoundingScale = %d: only the sales-prefixed key may be read", policy.RoundingScale)
	}
}
