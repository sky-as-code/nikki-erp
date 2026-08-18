package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// Rounding to a currency's own precision is the whole point of exposing decimal_places: an amount
// carrying more fractional digits than the currency is quoted to cannot be paid, and two modules
// rounding it differently produce totals that will not reconcile.
func TestRoundToCurrencyPrecision(t *testing.T) {
	testCases := []struct {
		name   string
		amount string
		places int32
		want   string
	}{
		{"VND is quoted in whole dong", "1234.56", 0, "1235"},
		{"USD keeps cents", "1234.567", 2, "1234.57"},
		{"KWD keeps three places", "1.23456", 3, "1.235"},
		{"an exact amount is unchanged", "1234.50", 2, "1234.5"},
		{"rounds half away from zero", "0.125", 2, "0.13"},
		{"a negative amount rounds away from zero too", "-0.125", 2, "-0.13"},
		{"a whole amount survives zero places", "2000000", 0, "2000000"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := decimal.RequireFromString(testCase.amount).Round(testCase.places)

			assert.True(t, got.Equal(decimal.RequireFromString(testCase.want)),
				"want %s, got %s", testCase.want, got)
		})
	}
}

// A currency row that carries no decimal_places must not be read as zero.
//
// Zero is a legitimate precision — VND is quoted in whole dong — so falling back to the zero value
// would be indistinguishable from a real VND row and would silently round the cents off a USD
// amount. The schema makes the column required with a default of 2, so this is only reachable
// through a migration that wrote a row directly.
func TestDecimalPlacesFallsBackToTwoWhenAbsent(t *testing.T) {
	withoutPlaces := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldCode: "USD",
	})

	assert.Equal(t, int32(defaultDecimalPlaces), decimalPlacesOf(withoutPlaces))
}

// Zero must survive as zero, which is the case the fallback above must not swallow.
func TestDecimalPlacesKeepsAnExplicitZero(t *testing.T) {
	vnd := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldCode:          "VND",
		models.CurrencyFieldDecimalPlaces: int32(0),
	})

	assert.Equal(t, int32(0), decimalPlacesOf(vnd))
}

func TestDecimalPlacesReadsAnExplicitValue(t *testing.T) {
	kwd := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldCode:          "KWD",
		models.CurrencyFieldDecimalPlaces: int32(3),
	})

	assert.Equal(t, int32(3), decimalPlacesOf(kwd))
}

// toResultData is what crosses the module boundary, so the fields a consumer relies on to validate
// a reference must all survive the mapping — in particular is_active and is_archived, which answer
// different questions and are both needed by AssertUsable.
func TestToResultDataCarriesWhatAConsumerNeeds(t *testing.T) {
	src := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldId:            "01JCURRENCYUSD00000000000",
		models.CurrencyFieldCode:          "USD",
		models.CurrencyFieldSymbol:        "$",
		models.CurrencyFieldDecimalPlaces: int32(2),
		models.CurrencyFieldIsActive:      true,
	})

	got := toResultData(src)

	assert.Equal(t, "01JCURRENCYUSD00000000000", string(got.Id))
	assert.Equal(t, "USD", got.Code)
	assert.Equal(t, "$", got.Symbol)
	assert.Equal(t, int32(2), got.DecimalPlaces)
	assert.True(t, got.IsActive)
	assert.False(t, got.IsArchived)
}

// An inactive currency is not an archived one. A consumer that collapsed the two would either
// refuse to read historical amounts or let a withdrawn currency into new records.
func TestToResultDataDistinguishesInactiveFromArchived(t *testing.T) {
	inactive := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldId:       "01JCURRENCYOLD00000000000",
		models.CurrencyFieldIsActive: false,
	})

	got := toResultData(inactive)

	assert.False(t, got.IsActive, "an inactive currency must report itself inactive")
	assert.False(t, got.IsArchived, "inactive alone must not read as archived")
}

// The zero value of a missing is_active must be false: a row with no flag is not usable by default.
func TestToResultDataTreatsMissingIsActiveAsInactive(t *testing.T) {
	src := models.NewCurrencyFrom(dmodel.DynamicFields{
		models.CurrencyFieldId: "01JCURRENCYNEW00000000000",
	})

	assert.False(t, util.ValueOrZeroOf(src.GetIsActive()))
	assert.False(t, toResultData(src).IsActive)
}
