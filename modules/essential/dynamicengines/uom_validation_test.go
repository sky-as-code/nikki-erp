package dynamicengines

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// BR-UOM-ESS-006 and BR-UOM-ESS-009: the conversion factor must agree with the UoM type.
func TestAssertFactorMatchesUomType(t *testing.T) {
	testCases := []struct {
		name       string
		uomType    models.UomType
		factor     string
		wantErrKey string
	}{
		{"reference with factor 1 is valid", models.UomTypeReference, "1", ""},
		{"reference with factor 1.000 is valid", models.UomTypeReference, "1.000", ""},
		{"reference with any other factor is rejected", models.UomTypeReference, "1000",
			"uom.reference_factor_must_be_one"},
		{"bigger_equal at exactly 1 is valid", models.UomTypeBiggerEqual, "1", ""},
		{"bigger_equal above 1 is valid", models.UomTypeBiggerEqual, "1000", ""},
		{"bigger_equal below 1 is rejected", models.UomTypeBiggerEqual, "0.001",
			"uom.bigger_equal_factor_out_of_range"},
		{"smaller below 1 is valid", models.UomTypeSmaller, "0.001", ""},
		{"smaller at exactly 1 is rejected", models.UomTypeSmaller, "1",
			"uom.smaller_factor_out_of_range"},
		{"smaller at zero is rejected", models.UomTypeSmaller, "0",
			"uom.smaller_factor_out_of_range"},
		{"smaller when negative is rejected", models.UomTypeSmaller, "-0.5",
			"uom.smaller_factor_out_of_range"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := &ft.ClientErrors{}

			assertFactorMatchesUomType(newUomFixture(t, testCase.uomType, testCase.factor, "0.01"), vErrs)

			assertViolation(t, vErrs, testCase.wantErrKey)
		})
	}
}

// BR-UOM-ESS-017: 0 <= rounding <= 1. The upper bound is inclusive because a step of
// exactly 1 is the "whole units only" precision of a discrete UoM such as Unit or gram.
func TestAssertRoundingInRange(t *testing.T) {
	testCases := []struct {
		name       string
		rounding   string
		wantErrKey string
	}{
		{"zero is valid", "0", ""},
		{"a fraction is valid", "0.01", ""},
		{"just below one is valid", "0.999999", ""},
		{"exactly one is valid", "1", ""},
		{"above one is rejected", "1.5", "uom.rounding_out_of_range"},
		{"negative is rejected", "-0.01", "uom.rounding_out_of_range"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			vErrs := &ft.ClientErrors{}

			assertRoundingInRange(newUomFixture(t, models.UomTypeBiggerEqual, "1", testCase.rounding), vErrs)

			assertViolation(t, vErrs, testCase.wantErrKey)
		})
	}
}

// A partial update carries only the changed fields, so the invariants must be checked
// against the record the update produces, not against the submitted fragment alone.
func TestMergeUomForValidationOverlaysSubmittedFields(t *testing.T) {
	stored := newUomFixture(t, models.UomTypeBiggerEqual, "1000", "0.01")
	submitted := models.NewUomFrom(dmodel.DynamicFields{})
	submitted.SetFactor(decimalPtr(t, "0.001"))

	merged := mergeUomForValidation(submitted, stored)

	require.NotNil(t, merged.GetFactor())
	assert.True(t, merged.GetFactor().Equal(decimal.RequireFromString("0.001")),
		"submitted factor wins")
	require.NotNil(t, merged.GetUomType())
	assert.Equal(t, models.UomTypeBiggerEqual, *merged.GetUomType(),
		"unsubmitted fields fall back to the stored record")
}

// Merging must not mutate the stored record it reads from.
func TestMergeUomForValidationLeavesStoredUntouched(t *testing.T) {
	stored := newUomFixture(t, models.UomTypeBiggerEqual, "1000", "0.01")
	submitted := models.NewUomFrom(dmodel.DynamicFields{})
	submitted.SetUomType(models.WrapUomType(models.UomTypeSmaller.String()))

	mergeUomForValidation(submitted, stored)

	require.NotNil(t, stored.GetUomType())
	assert.Equal(t, models.UomTypeBiggerEqual, *stored.GetUomType())
}

// BR-UOM-ESS-020: factor, type and category freeze once the UoM is used by transactions.
// isUomInUse is stubbed to false until a consuming module exists, so nothing is rejected
// yet — this pins which fields the rule guards when the probe goes live.
func TestAssertImmutableWhileInUseIsInertWithoutConsumers(t *testing.T) {
	stored := newUomFixture(t, models.UomTypeBiggerEqual, "1000", "0.01")
	params := dmodel.DynamicFields{models.UomFieldFactor: "24"}
	vErrs := &ft.ClientErrors{}

	assertImmutableWhileInUse(nil, params, stored, vErrs)

	assert.Zero(t, vErrs.Count(), "no consuming module reports usage yet")
}

func newUomFixture(t *testing.T, uomType models.UomType, factor string, rounding string) *models.Uom {
	t.Helper()
	uom := models.NewUom()
	uom.SetUomType(models.WrapUomType(uomType.String()))
	uom.SetFactor(decimalPtr(t, factor))
	uom.SetRounding(decimalPtr(t, rounding))
	return uom
}

func decimalPtr(t *testing.T, raw string) *decimal.Decimal {
	t.Helper()
	parsed, err := decimal.NewFromString(raw)
	require.NoError(t, err)
	return &parsed
}

func assertViolation(t *testing.T, vErrs *ft.ClientErrors, wantErrKey string) {
	t.Helper()
	if wantErrKey == "" {
		assert.Zero(t, vErrs.Count(), "expected no violation")
		return
	}
	require.Equal(t, 1, vErrs.Count(), "expected exactly one violation")
	assert.Equal(t, wantErrKey, (*vErrs)[0].Key)
}
