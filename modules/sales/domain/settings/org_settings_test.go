package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A settings schema describes values; it must not own a table. PrimaryKeys is populated only by
// populateDbMetadata, which runs only under ShouldBuildDb, so an empty one is the check. The
// settings module refuses a registration that fails this, so getting it wrong fails the boot.
func TestOrgSettingsSchema_IsMetadataOnly(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Empty(t, schema.PrimaryKeys(),
		"a settings schema must not declare should_build_db: settings owns the only tables")
	assert.Empty(t, schema.TableName())
}

// The schema name is the contract with the settings module and with the frontend, which both key
// off the string rather than off this package.
func TestOrgSettingsSchema_Name(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Equal(t, OrgSettingsSchemaName, schema.Name())
	assert.Equal(t, "sales_org_settings", schema.Name())
}

// Every setting the plan names must be present. A missing one is not a compile error anywhere —
// the constant would still exist and the reader would silently fall back to its default forever.
func TestOrgSettingsSchema_DeclaresEverySetting(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	for _, name := range []string{
		OrgSettingMaxPaymentMethodsPerBill,
		OrgSettingReturnWindowDays,
		OrgSettingAllowOverpayment,
		OrgSettingAllowCashChange,
		OrgSettingDraftOrderExpiryHours,
		OrgSettingRoundingScale,
		OrgSettingDefaultTaxRate,
		OrgSettingDefaultSalesTaxCode,
	} {
		_, ok := schema.Field(name)
		assert.Truef(t, ok, "the schema declares no %q field, so nothing can ever configure it", name)
	}
	assert.Len(t, schema.Fields(), 8, "a field added without a constant is unreachable by name")
}

// The settings that change what money MEANS must not be overridable.
//
// Absent metadata reads as overridable, so a dropped entry would not fail anything — it would
// quietly let two organizations round to different scales, producing totals that cannot be added
// together in a consolidated report, or apply different tax rates to the same product, producing
// fiscal documents that disagree about what was owed.
func TestOrgSettingsSchema_MonetarySettingsForbidOverride(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	for _, name := range []string{
		OrgSettingRoundingScale,
		OrgSettingDefaultTaxRate,
		OrgSettingDefaultSalesTaxCode,
	} {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		allow, found := field.MetadataValue("allow_override")
		assert.Truef(t, found, "%s must declare allow_override explicitly", name)
		assert.Equalf(t, false, allow,
			"%s changes what money means; two organizations disagreeing on it produce numbers "+
				"that cannot be reconciled", name)
	}
}

// The five genuine per-organization policies must stay overridable, and must say so explicitly
// rather than relying on the absent-means-true default — an explicit true is a decision, an
// omission is an oversight, and they read identically at runtime.
func TestOrgSettingsSchema_PolicySettingsAllowOverride(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	for _, name := range []string{
		OrgSettingMaxPaymentMethodsPerBill,
		OrgSettingReturnWindowDays,
		OrgSettingAllowOverpayment,
		OrgSettingAllowCashChange,
		OrgSettingDraftOrderExpiryHours,
	} {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		allow, found := field.MetadataValue("allow_override")
		assert.Truef(t, found, "%s must declare allow_override explicitly", name)
		assert.Equalf(t, true, allow,
			"%s is a commercial policy a business unit sets for its own trading", name)
	}
}

// The Go defaults and the JSON defaults must agree.
//
// They are deliberately duplicated: the JSON default governs what an administrator sees in the
// settings UI, and the Go one governs what the code does when the settings read fails. That is a
// real need, but it is also two copies of one number — so this test is what stops them drifting.
//
// rounding_scale and default_tax_rate are absent by design and are checked separately below.
func TestDefaultsAgreeWithSchema(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	// The JSON default arrives as whatever encoding/json chose, so a number is a float64 however it
	// was written. The comparison is therefore against that shape rather than against int32 - which
	// is also exactly the conversion ResolveSalesPolicy has to make at runtime.
	cases := map[string]any{
		OrgSettingMaxPaymentMethodsPerBill: float64(DefaultMaxPaymentMethodsPerBill),
		OrgSettingReturnWindowDays:         float64(DefaultReturnWindowDays),
		OrgSettingAllowOverpayment:         DefaultAllowOverpayment,
		OrgSettingAllowCashChange:          DefaultAllowCashChange,
		OrgSettingDraftOrderExpiryHours:    float64(DefaultDraftOrderExpiryHours),
	}
	for name, want := range cases {
		field, ok := schema.Field(name)
		require.True(t, ok, name)

		declared := field.Default()
		require.NotNilf(t, declared, "%s declares no default_value in the JSON", name)
		assert.Truef(t, declared.Same(want),
			"the Go default for %s and its default_value in org_settings.json disagree: "+
				"schema has %v, Go has %v", name, *declared.Get(), want)
	}
}

// rounding_scale deliberately carries NO default_value, and its floor is 1 rather than 0.
//
// ModelField.Validate treats a numeric zero as an ABSENT value platform-wide, so a declared
// default of 0 would read back as "unset" and a min of 0 would reject nothing. The enforceable
// range therefore starts at 1, and zero — whole-dong rounding for VND — lives in Go as
// DefaultRoundingScale, applied when the setting is unset. This test pins that reasoning so a
// later edit "fixing" the missing default reintroduces the trap knowingly.
func TestRoundingScaleHasNoDefaultAndAFloorOfOne(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	field, ok := schema.Field(OrgSettingRoundingScale)
	require.True(t, ok)

	assert.Nil(t, field.Default(),
		"a default of 0 would read back as unset: numeric zero is treated as absent platform-wide")
	assert.Equal(t, int32(0), DefaultRoundingScale,
		"VND has no minor unit, so an unconfigured organization rounds to whole dong")

	bounds, ok := field.DataType().Options()[dmodel.FieldDataTypeOptRange].([]int32)
	require.True(t, ok, "rounding_scale must be an int32 with a range")
	require.Len(t, bounds, 2)
	assert.Equal(t, int32(1), bounds[0],
		"the floor is 1 because a floor of 0 could not reject 0 anyway")
	assert.Equal(t, int32(4), bounds[1],
		"four places matches the internal monetary scale of D-01")
}

// default_tax_rate carries no default either, and that is D-38 rather than an oversight: there is
// no tax master anywhere in the codebase, so the resolver returns zero and this setting exists as
// the manual override for an organization that must charge a flat rate before one is built.
func TestDefaultTaxRateHasNoDefaultAndIsAFraction(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	field, ok := schema.Field(OrgSettingDefaultTaxRate)
	require.True(t, ok)

	assert.Nil(t, field.Default(),
		"an organization that has not set a rate has not chosen zero; the resolver supplies it")
	assert.True(t, DefaultTaxRate.IsZero(),
		"D-38: the rate resolves to zero until a tax master exists")
}
