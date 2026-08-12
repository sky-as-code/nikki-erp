package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


// BR-UOM-ESS-015: rounding precision is a step, not a digit count.
func TestApplyRounding(t *testing.T) {
	testCases := []struct {
		name     string
		quantity string
		rounding string
		want     string
	}{
		{"a zero rounding keeps the exact value", "0.123456", "0", "0.123456"},
		{"two decimal places", "1.2345", "0.01", "1.23"},
		{"rounds half away from zero", "1.235", "0.01", "1.24"},
		{"whole units", "2.4", "1", "2"},
		{"whole units rounding up", "2.6", "1", "3"},
		{"a non-decimal step snaps to quarters", "1.30", "0.25", "1.25"},
		{"a non-decimal step snaps upward", "1.40", "0.25", "1.50"},
		{"an exact value is unchanged", "2000000", "1", "2000000"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			rounding := decimal.RequireFromString(testCase.rounding)

			got := applyRounding(decimal.RequireFromString(testCase.quantity), &rounding)

			assert.True(t, got.Equal(decimal.RequireFromString(testCase.want)),
				"want %s, got %s", testCase.want, got)
		})
	}
}

func TestApplyRoundingWithoutPrecisionKeepsExactValue(t *testing.T) {
	quantity := decimal.RequireFromString("0.9144")

	assert.True(t, applyRounding(quantity, nil).Equal(quantity))
}

// BR-UOM-ESS-013's worked example: with kg as the reference, 2 ton converts to 2,000,000 g.
func TestConversionFormulaWorkedExample(t *testing.T) {
	quantity := decimal.RequireFromString("2")
	tonFactor := decimal.RequireFromString("1000")
	gramFactor := decimal.RequireFromString("0.001")

	got := quantity.Mul(tonFactor).DivRound(gramFactor, conversionScale)

	assert.True(t, got.Equal(decimal.RequireFromString("2000000")), "got %s", got)
}

// BR-UOM-ESS-014: a chained Carton -> Box -> Unit relationship resolves through the
// reference UoM in a single step, with no hierarchy stored.
func TestChainedConversionThroughReference(t *testing.T) {
	// Unit is the reference (factor 1); a Box is 12 Units; a Carton is 144 Units.
	boxFactor := decimal.RequireFromString("12")
	cartonFactor := decimal.RequireFromString("144")
	unitFactor := decimal.RequireFromString("1")
	twoCartons := decimal.RequireFromString("2")

	inBoxes := twoCartons.Mul(cartonFactor).DivRound(boxFactor, conversionScale)
	inUnits := twoCartons.Mul(cartonFactor).DivRound(unitFactor, conversionScale)

	assert.True(t, inBoxes.Equal(decimal.RequireFromString("24")), "got %s boxes", inBoxes)
	assert.True(t, inUnits.Equal(decimal.RequireFromString("288")), "got %s units", inUnits)
}

// BR-UOM-ESS-018: intermediate precision must survive a round trip that float64 would not.
func TestConversionKeepsIntermediatePrecision(t *testing.T) {
	yardFactor := decimal.RequireFromString("0.9144")
	meterFactor := decimal.RequireFromString("1")
	oneYard := decimal.RequireFromString("1")

	inMeters := oneYard.Mul(yardFactor).DivRound(meterFactor, conversionScale)
	backToYards := inMeters.Mul(meterFactor).DivRound(yardFactor, conversionScale)

	require.True(t, inMeters.Equal(decimal.RequireFromString("0.9144")), "got %s", inMeters)
	assert.True(t, backToYards.Equal(oneYard), "round trip drifted to %s", backToYards)
}
