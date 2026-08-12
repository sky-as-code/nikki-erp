package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func dec(t *testing.T, value string) decimal.Decimal {
	t.Helper()
	parsed, err := decimal.NewFromString(value)
	assert.NoError(t, err)
	return parsed
}

func TestAvailableQuantitySubtractsReserved(t *testing.T) {
	onHand := dec(t, "10.5")
	reserved := dec(t, "4.25")

	assert.True(t, AvailableQuantity(&onHand, &reserved).Equal(dec(t, "6.25")))
}

// A missing operand means "none recorded", which is zero. Treating it as unknown and refusing to
// compute would leave the field blank on every quant that has never been reserved against.
func TestAvailableQuantityTreatsNilAsZero(t *testing.T) {
	onHand := dec(t, "7")

	assert.True(t, AvailableQuantity(&onHand, nil).Equal(dec(t, "7")))
	assert.True(t, AvailableQuantity(nil, &onHand).Equal(dec(t, "-7")))
	assert.True(t, AvailableQuantity(nil, nil).IsZero())
}

// Reserving everything leaves nothing available, which is zero rather than an absent value.
func TestAvailableQuantityIsZeroWhenFullyReserved(t *testing.T) {
	quantity := dec(t, "3")

	assert.True(t, AvailableQuantity(&quantity, &quantity).IsZero())
}

// The precision the BR requires: a decimal that a float64 cannot hold exactly must survive the
// subtraction unchanged, which is why quantities are decimals rather than floats.
func TestAvailableQuantityKeepsDecimalPrecision(t *testing.T) {
	onHand := dec(t, "0.3")
	reserved := dec(t, "0.1")

	assert.Equal(t, "0.2", AvailableQuantity(&onHand, &reserved).String())
}

func TestFillAvailableQuantityWritesTheDerivedField(t *testing.T) {
	onHand := dec(t, "12")
	reserved := dec(t, "5")
	quant := models.NewStockQuant()
	quant.SetOnHandQuantity(&onHand)
	quant.SetReservedQuantity(&reserved)

	FillAvailableQuantity(quant.GetFieldData())

	filled := quant.GetAvailableQuantity()
	assert.NotNil(t, filled)
	assert.True(t, filled.Equal(dec(t, "7")))
}

func TestFillAvailableQuantityOnEmptyRowIsZero(t *testing.T) {
	fields := dmodel.DynamicFields{}

	FillAvailableQuantity(fields)

	filled := models.NewStockQuantFrom(fields).GetAvailableQuantity()
	assert.NotNil(t, filled)
	assert.True(t, filled.IsZero())
}

func TestAssertQuantNotClientWritableReportsABusinessViolation(t *testing.T) {
	vErrs := ft.NewClientErrors()

	AssertQuantNotClientWritable(vErrs)

	assert.Equal(t, 1, vErrs.Count())
}

// An empty projection means "the default field set", which includes the derived quantity.
func TestWantsAvailableQuantity(t *testing.T) {
	assert.True(t, wantsAvailableQuantity(nil))
	assert.True(t, wantsAvailableQuantity([]string{}))
	assert.True(t, wantsAvailableQuantity([]string{models.StockQuantFieldAvailableQuantity}))
	assert.False(t, wantsAvailableQuantity([]string{models.StockQuantFieldOnHandQuantity}))
}

// The operands must be fetched, or the arithmetic has nothing to work from.
func TestAvailableQuantityOperands(t *testing.T) {
	assert.ElementsMatch(t,
		[]string{models.StockQuantFieldOnHandQuantity, models.StockQuantFieldReservedQuantity},
		availableQuantityOperands())
}
