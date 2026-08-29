package services

import (
	"fmt"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
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

// A decimal a float64 cannot hold exactly must survive the subtraction unchanged, which is why
// quantities are decimals rather than floats.
func TestAvailableQuantityKeepsDecimalPrecision(t *testing.T) {
	onHand := dec(t, "0.3")
	reserved := dec(t, "0.1")

	assert.Equal(t, "0.2", AvailableQuantity(&onHand, &reserved).String())
}

// available_quantity is a computed field in the schema, so this pins the declared formula to the
// pure rule above: evaluating the schema's expression over a row must match AvailableQuantity.
func TestComputedDefinitionMatchesAvailableQuantityRule(t *testing.T) {
	schema := models.StockQuantSchemaBuilder().Build()
	field, ok := schema.Field(models.StockQuantFieldAvailableQuantity)
	assert.True(t, ok)
	assert.True(t, field.IsComputed())
	assert.True(t, field.IsVirtual())

	cases := []struct {
		name             string
		onHand, reserved *decimal.Decimal
	}{
		{"both set", ptrDec(t, "10.5"), ptrDec(t, "4.25")},
		{"nil reserved reads as zero", ptrDec(t, "7"), nil},
		{"nil on-hand reads as zero", nil, ptrDec(t, "7")},
		{"both nil is zero", nil, nil},
		{"decimal precision survives", ptrDec(t, "0.3"), ptrDec(t, "0.1")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			row := dmodel.DynamicFields{
				models.StockQuantFieldOnHandQuantity:   tc.onHand,
				models.StockQuantFieldReservedQuantity: tc.reserved,
			}
			def, err := computed.DefOf(field)
			assert.NoError(t, err)
			got, err := computed.Eval(def.Expression, row)
			assert.NoError(t, err)

			want := AvailableQuantity(tc.onHand, tc.reserved)
			gotDec, convErr := decimal.NewFromString(fmt.Sprint(got))
			assert.NoError(t, convErr)
			assert.True(t, want.Equal(gotDec), "want %s, got %v", want, got)
		})
	}
}

func ptrDec(t *testing.T, value string) *decimal.Decimal {
	parsed := dec(t, value)
	return &parsed
}

func TestAssertQuantNotClientWritableReportsABusinessViolation(t *testing.T) {
	vErrs := ft.NewClientErrors()

	AssertQuantNotClientWritable(vErrs)

	assert.Equal(t, 1, vErrs.Count())
}
