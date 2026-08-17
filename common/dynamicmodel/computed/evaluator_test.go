package computed

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func mustEval(t *testing.T, expr Expr, row dmodel.DynamicFields) any {
	t.Helper()
	got, err := Eval(expr, row)
	require.NoError(t, err)
	return got
}

func TestEval_AvailableQuantityFormula(t *testing.T) {
	// The stock-quant archetype: coalesce turns an unwritten operand into zero explicitly,
	// because the framework itself never converts nil to zero.
	formula := Sub(
		Fn("coalesce", F("on_hand_quantity"), Lit(decimal.Zero)),
		Fn("coalesce", F("reserved_quantity"), Lit(decimal.Zero)),
	)

	onHand := decimal.NewFromInt(10)
	reserved := decimal.NewFromInt(3)
	got := mustEval(t, formula, dmodel.DynamicFields{
		"on_hand_quantity": &onHand, "reserved_quantity": &reserved,
	})
	assert.True(t, decimal.NewFromInt(7).Equal(got.(decimal.Decimal)))

	got = mustEval(t, formula, dmodel.DynamicFields{"on_hand_quantity": onHand})
	assert.True(t, decimal.NewFromInt(10).Equal(got.(decimal.Decimal)))
}

func TestEval_NullPropagation(t *testing.T) {
	row := dmodel.DynamicFields{"qty": int64(5), "price": nil}

	cases := map[string]Expr{
		"multiply by null":  Mul(F("qty"), F("price")),
		"add missing field": Add(F("qty"), F("absent")),
		"compare with null": Gt(F("qty"), F("price")),
		"lower of null":     Fn("lower", F("price")),
		"concat with null":  Fn("concat", Lit("a"), F("price")),
		"negate null":       Neg(F("price")),
		"round null":        Fn("round", F("price")),
		"date_add null":     Fn("date_add", F("price"), Lit(1)),
		"typed nil pointer": Mul(F("qty"), F("typed_nil")),
	}
	row["typed_nil"] = (*decimal.Decimal)(nil)

	for name, expr := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Nil(t, mustEval(t, expr, row))
		})
	}
}

func TestEval_ThreeValuedLogic(t *testing.T) {
	row := dmodel.DynamicFields{"t": true, "f": false, "n": nil}

	assert.Equal(t, false, mustEval(t, And(F("f"), F("n")), row))
	assert.Nil(t, mustEval(t, And(F("t"), F("n")), row))
	assert.Equal(t, true, mustEval(t, Or(F("t"), F("n")), row))
	assert.Nil(t, mustEval(t, Or(F("f"), F("n")), row))
	assert.Nil(t, mustEval(t, Not(F("n")), row))
	assert.Equal(t, false, mustEval(t, Not(F("t")), row))
	assert.Equal(t, true, mustEval(t, IsNull(F("n")), row))
	assert.Equal(t, false, mustEval(t, IsNotNull(F("n")), row))
	assert.Equal(t, true, mustEval(t, IsNotNull(F("t")), row))
}

func TestEval_ArithmeticKinds(t *testing.T) {
	row := dmodel.DynamicFields{
		"i":   int64(7),
		"j":   int32(2),
		"dec": decimal.NewFromFloat(2.5),
	}

	assert.Equal(t, int64(9), mustEval(t, Add(F("i"), F("j")), row))
	assert.Equal(t, int64(14), mustEval(t, Mul(F("i"), F("j")), row))
	assert.Equal(t, int64(1), mustEval(t, Mod(F("i"), F("j")), row))

	div := mustEval(t, Div(F("i"), F("j")), row)
	assert.True(t, decimal.NewFromFloat(3.5).Equal(div.(decimal.Decimal)))

	mixed := mustEval(t, Mul(F("j"), F("dec")), row)
	assert.True(t, decimal.NewFromInt(5).Equal(mixed.(decimal.Decimal)))

	_, err := Eval(Div(F("i"), Lit(0)), row)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "division by zero")
}

func TestEval_ComparisonFamilies(t *testing.T) {
	earlier := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	later := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	row := dmodel.DynamicFields{
		"a": decimal.NewFromInt(5), "b": int64(5),
		"s1": "apple", "s2": "banana",
		"d1": earlier, "d2": later,
	}

	assert.Equal(t, true, mustEval(t, Eq(F("a"), F("b")), row))
	assert.Equal(t, true, mustEval(t, Lt(F("s1"), F("s2")), row))
	assert.Equal(t, true, mustEval(t, Lt(F("d1"), F("d2")), row))
	assert.Equal(t, false, mustEval(t, Ge(F("d1"), F("d2")), row))
}

func TestEval_CasePicksFirstTrueBranch(t *testing.T) {
	stockStatus := Case().
		When(Le(F("available_qty"), Lit(0)), Lit("out_of_stock")).
		When(Lt(F("available_qty"), F("reorder_point")), Lit("low_stock")).
		Else(Lit("in_stock"))

	assert.Equal(t, "out_of_stock", mustEval(t, stockStatus, dmodel.DynamicFields{
		"available_qty": int64(0), "reorder_point": int64(5),
	}))
	assert.Equal(t, "low_stock", mustEval(t, stockStatus, dmodel.DynamicFields{
		"available_qty": int64(3), "reorder_point": int64(5),
	}))
	assert.Equal(t, "in_stock", mustEval(t, stockStatus, dmodel.DynamicFields{
		"available_qty": int64(9), "reorder_point": int64(5),
	}))
	// A nil condition is "not matched", like SQL CASE, so it falls through to ELSE.
	assert.Equal(t, "in_stock", mustEval(t, stockStatus, dmodel.DynamicFields{
		"available_qty": nil, "reorder_point": int64(5),
	}))
}

func TestEval_RelatedExprIsNotRowEvaluable(t *testing.T) {
	_, err := Eval(Related("template.name"), dmodel.DynamicFields{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "eval plan")
}
