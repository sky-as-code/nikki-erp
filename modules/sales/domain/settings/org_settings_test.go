package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// A settings schema describes values and must not own a table. PrimaryKeys is populated only under
// ShouldBuildDb, so an empty one is the check.
func TestOrgSettingsSchema_IsMetadataOnly(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Empty(t, schema.PrimaryKeys(),
		"a settings schema must not declare should_build_db: settings owns the only tables")
	assert.Empty(t, schema.TableName())
}

// The schema name is the contract with the settings module and the frontend, which key off the
// string rather than this package.
func TestOrgSettingsSchema_Name(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	assert.Equal(t, OrgSettingsSchemaName, schema.Name())
	assert.Equal(t, "sales_org_settings", schema.Name())
}

// Every declared setting must be present: a missing one is no compile error, and the reader would
// silently fall back to its default forever.
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
		OrgSettingOutgoingOperationTypeId,
		OrgSettingIncomingOperationTypeId,
	} {
		_, ok := schema.Field(name)
		assert.Truef(t, ok, "the schema declares no %q field, so nothing can ever configure it", name)
	}
	assert.Len(t, schema.Fields(), 10, "a field added without a constant is unreachable by name")
}

// The settings that change what money MEANS must not be overridable. Absent metadata reads as
// overridable, so a dropped entry would fail nothing and quietly let two organizations round to
// different scales or tax the same product differently.
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

// The genuine per-organization policies must stay overridable and say so explicitly: an omission
// and a deliberate true read identically at runtime.
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

// The Go defaults and the JSON defaults are deliberately duplicated (UI vs. read-failure fallback);
// this test stops the two copies drifting. rounding_scale and default_tax_rate are absent by design
// and checked separately below.
func TestDefaultsAgreeWithSchema(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	// The JSON default arrives as a float64 however it was written, so the comparison is against that
	// shape — the same conversion ResolveSalesPolicy makes at runtime.
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

// rounding_scale deliberately carries NO default_value and its floor is 1: ModelField.Validate
// treats a numeric zero as ABSENT platform-wide, so a declared 0 would read back as unset and a min
// of 0 would reject nothing. Zero lives in Go as DefaultRoundingScale. This test pins that so a
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

// default_tax_rate carries no default either, deliberately: the resolver returns zero and the
// setting survives only as a manual flat-rate override.
func TestDefaultTaxRateHasNoDefaultAndIsAFraction(t *testing.T) {
	schema := OrgSettingsSchemaBuilder().Build()

	field, ok := schema.Field(OrgSettingDefaultTaxRate)
	require.True(t, ok)

	assert.Nil(t, field.Default(),
		"an organization that has not set a rate has not chosen zero; the resolver supplies it")
	assert.True(t, DefaultTaxRate.IsZero(),
		"D-38: the rate resolves to zero until a tax master exists")
}
