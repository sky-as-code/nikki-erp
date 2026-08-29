package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The pricing service's own decisions. Ranking lives in vendorpricing and is tested there; here it
// is the three judgements the pure resolver cannot make: whether a quote's window is open, whether
// the caller stated a price of its own, and how a quantity is expressed in each quote's unit.

func atTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

func windowRow(from, to *time.Time) dmodel.DynamicFields {
	row := dmodel.DynamicFields{}
	if from != nil {
		value := model.ModelDateTime(*from)
		row[models.VendorProductPriceFieldValidFrom] = &value
	}
	if to != nil {
		value := model.ModelDateTime(*to)
		row[models.VendorProductPriceFieldValidTo] = &value
	}
	return row
}

func timePtr(t time.Time) *time.Time { return &t }

// A quote with neither bound is a standing offer. Reading an absent bound as closed would make
// every open-ended price invisible, and those are the most common rows in a vendor's list.
func TestAQuoteWithNoBoundsIsAlwaysOpen(t *testing.T) {
	assert.True(t, windowCovers(windowRow(nil, nil), atTime(2026, time.August, 28)))
}

func TestAQuoteIsClosedBeforeItsStartAndAfterItsEnd(t *testing.T) {
	row := windowRow(timePtr(atTime(2026, time.March, 1)), timePtr(atTime(2026, time.June, 30)))

	assert.False(t, windowCovers(row, atTime(2026, time.February, 1)), "not yet started")
	assert.True(t, windowCovers(row, atTime(2026, time.April, 15)), "inside the window")
	assert.False(t, windowCovers(row, atTime(2026, time.August, 1)), "expired")
}

// The bounds are inclusive: a quote valid "until the 30th" is usable on the 30th.
func TestTheWindowBoundsAreInclusive(t *testing.T) {
	day := atTime(2026, time.June, 30)
	row := windowRow(timePtr(atTime(2026, time.March, 1)), timePtr(day))

	assert.True(t, windowCovers(row, day), "the last day of validity is a day of validity")

	start := atTime(2026, time.March, 1)
	assert.True(t, windowCovers(windowRow(timePtr(start), nil), start),
		"the first day of validity is a day of validity")
}

// One open bound is open-ended in that direction: from March with no end is live forever, and an
// end with no start was live from the beginning.
func TestOneOpenBoundLeavesThatDirectionUnbounded(t *testing.T) {
	fromOnly := windowRow(timePtr(atTime(2026, time.March, 1)), nil)
	assert.False(t, windowCovers(fromOnly, atTime(2026, time.January, 1)))
	assert.True(t, windowCovers(fromOnly, atTime(2030, time.January, 1)))

	toOnly := windowRow(nil, timePtr(atTime(2026, time.June, 30)))
	assert.True(t, windowCovers(toOnly, atTime(2000, time.January, 1)))
	assert.False(t, windowCovers(toOnly, atTime(2026, time.July, 1)))
}

// A line arriving with a price is a negotiation and must survive; one arriving without is asking to
// be priced.
func TestAnExplicitPriceIsRecognisedByPresenceNotByValue(t *testing.T) {
	assert.False(t, hasExplicitPrice(dmodel.DynamicFields{}),
		"a line that names no price is asking to be priced")
	assert.False(t, hasExplicitPrice(dmodel.DynamicFields{
		models.PurchaseOrderLineFieldUnitPrice: nil,
	}), "an explicit null is not a price")
	assert.True(t, hasExplicitPrice(dmodel.DynamicFields{
		models.PurchaseOrderLineFieldUnitPrice: decimal.NewFromInt(9200),
	}))
}

// Zero is a real price: a free-of-charge line must not be overwritten with the vendor's list price,
// so the check tests presence rather than truthiness.
func TestAZeroPriceCountsAsExplicit(t *testing.T) {
	assert.True(t, hasExplicitPrice(dmodel.DynamicFields{
		models.PurchaseOrderLineFieldUnitPrice: decimal.Zero,
	}), "a free-of-charge line is a deliberate price, not an absent one")
}

// A section or a note buys nothing, so pricing must neither touch it nor fail on it.
func TestPriceLineLeavesANonProductLineAlone(t *testing.T) {
	pricer := NewLinePricer(nil)
	line := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeSection),
	}

	require.NoError(t, pricer.PriceLine(nil, line, dmodel.DynamicFields{}, "01TEMPLATE",
		atTime(2026, time.August, 28)))
	assert.NotContains(t, line, models.PurchaseOrderLineFieldVendorProductPriceId)
}

// An order with no vendor, or a line with no product, cannot be priced from a vendor's price list.
// Both are ordinary states, so neither reads the database and neither is an error.
func TestPriceLineSkipsWhatCannotBePriced(t *testing.T) {
	pricer := NewLinePricer(nil)
	priced := atTime(2026, time.August, 28)

	noVendor := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldOrgId:    "01ORG",
	}
	require.NoError(t, pricer.PriceLine(nil, noVendor, dmodel.DynamicFields{}, "01TEMPLATE", priced))
	assert.NotContains(t, noVendor, models.PurchaseOrderLineFieldVendorProductPriceId)

	noProduct := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldOrgId:    "01ORG",
	}
	order := dmodel.DynamicFields{models.PurchaseOrderFieldVendorId: "01VENDOR"}
	require.NoError(t, pricer.PriceLine(nil, noProduct, order, "", priced))
	assert.NotContains(t, noProduct, models.PurchaseOrderLineFieldVendorProductPriceId)
}

// A line already counted in the quote's unit needs no conversion, so Essential is not called.
func TestConvertIsSkippedWhenTheUnitsAlreadyMatch(t *testing.T) {
	pricer := NewLinePricer(nil)

	got, ok, err := pricer.convert(nil, decimal.NewFromInt(12), "01UOM_CASE", "01UOM_CASE")

	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, got.Equal(decimal.NewFromInt(12)))
}

// A line with no unit carries a bare number, read as already in whatever the quote is per.
func TestALineWithNoUnitIsTakenAsAlreadyInTheQuotesUnit(t *testing.T) {
	pricer := NewLinePricer(nil)

	got, ok, err := pricer.convert(nil, decimal.NewFromInt(5), "", "01UOM_CASE")

	require.NoError(t, err)
	require.True(t, ok)
	assert.True(t, got.Equal(decimal.NewFromInt(5)))
}

// With no unit port bound there is no way to convert, so the unit is reported unavailable rather
// than as a zero, which would make every break in it look reached.
func TestAMissingUnitPortReportsUnavailableRatherThanZero(t *testing.T) {
	pricer := NewLinePricer(nil)

	_, ok, err := pricer.convert(nil, decimal.NewFromInt(12), "01UOM_CASE", "01UOM_PIECE")

	require.NoError(t, err)
	assert.False(t, ok, "no conversion means the candidate is skipped, not priced at a zero break")
}

// int32Of has to survive whatever shape the repository hands back: a lead time read as zero because
// it arrived as an int64 would silently understate every delivery estimate.
func TestInt32OfAcceptsTheShapesARepositoryReturns(t *testing.T) {
	value := int32(30)
	testCases := []struct {
		name  string
		value any
		want  int32
	}{
		{"int32", int32(30), 30},
		{"pointer", &value, 30},
		{"int", 30, 30},
		{"int64", int64(30), 30},
		{"float64", float64(30), 30},
		{"absent", nil, 0},
		{"unrecognised", "30", 0},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.want,
				int32Of(dmodel.DynamicFields{"lead_time_days": testCase.value}, "lead_time_days"))
		})
	}
}
