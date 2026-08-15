package computed

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func TestFunctions_StringFamily(t *testing.T) {
	row := dmodel.DynamicFields{"code": "PV-01", "name": "  Widget  "}

	assert.Equal(t, "pv-01", mustEval(t, Fn("lower", F("code")), row))
	assert.Equal(t, "PV-01", mustEval(t, Fn("upper", F("code")), row))
	assert.Equal(t, "Widget", mustEval(t, Fn("trim", F("name")), row))
	assert.Equal(t, int32(5), mustEval(t, Fn("length", F("code")), row))
	assert.Equal(t, "PV-01 - Widget",
		mustEval(t, Fn("concat", F("code"), Lit(" - "), Fn("trim", F("name"))), row))
	// Length counts runes, not bytes.
	assert.Equal(t, int32(3), mustEval(t, Fn("length", Lit("một")), dmodel.DynamicFields{}))
}

func TestFunctions_NumericFamily(t *testing.T) {
	row := dmodel.DynamicFields{
		"price": decimal.NewFromFloat(12.345),
		"count": int64(-4),
	}

	round := mustEval(t, Fn("round", F("price"), Lit(2)), row)
	assert.True(t, decimal.NewFromFloat(12.35).Equal(round.(decimal.Decimal)), round)

	ceil := mustEval(t, Fn("ceil", F("price")), row)
	assert.True(t, decimal.NewFromInt(13).Equal(ceil.(decimal.Decimal)))

	floor := mustEval(t, Fn("floor", F("price")), row)
	assert.True(t, decimal.NewFromInt(12).Equal(floor.(decimal.Decimal)))

	// Integer input keeps producing an integer.
	assert.Equal(t, int64(4), mustEval(t, Fn("abs", F("count")), row))
	assert.Equal(t, int64(-4), mustEval(t, Fn("round", F("count")), row))
}

func TestFunctions_NullFamily(t *testing.T) {
	row := dmodel.DynamicFields{"a": nil, "b": "fallback", "x": int64(5), "y": int64(5), "z": int64(6)}

	assert.Equal(t, "fallback", mustEval(t, Fn("coalesce", F("a"), F("b")), row))
	assert.Nil(t, mustEval(t, Fn("coalesce", F("a"), F("a")), row))
	assert.Nil(t, mustEval(t, Fn("nullif", F("x"), F("y")), row))
	assert.Equal(t, int64(5), mustEval(t, Fn("nullif", F("x"), F("z")), row))
	assert.Nil(t, mustEval(t, Fn("nullif", F("a"), F("x")), row))
}

func TestFunctions_DateTimeFamily(t *testing.T) {
	base := time.Date(2026, 8, 15, 10, 30, 0, 0, time.UTC)
	row := dmodel.DynamicFields{"due": base}

	added := mustEval(t, Fn("date_add", F("due"), Lit(10)), row).(time.Time)
	assert.Equal(t, time.Date(2026, 8, 25, 10, 30, 0, 0, time.UTC), added)

	diff := mustEval(t, Fn("date_diff", Lit(added), F("due")), row)
	assert.Equal(t, int32(10), diff)

	assert.Equal(t, int32(2026), mustEval(t, Fn("extract", Lit("year"), F("due")), row))
	assert.Equal(t, int32(8), mustEval(t, Fn("extract", Lit("month"), F("due")), row))
	assert.Equal(t, int32(227), mustEval(t, Fn("extract", Lit("doy"), F("due")), row))

	today := mustEval(t, Fn("today"), row).(time.Time)
	assert.Equal(t, 0, today.Hour())
	now := mustEval(t, Fn("now"), row).(time.Time)
	assert.False(t, now.IsZero())
}

func TestFunctions_ExtractRejectsUnknownPart(t *testing.T) {
	row := dmodel.DynamicFields{"due": time.Now()}

	_, err := Eval(Fn("extract", Lit("century"), F("due")), row)
	require.Error(t, err)
	assert.False(t, IsValidExtractPart("century"))
	assert.True(t, IsValidExtractPart("year"))
}

func TestFunctions_LookupAndArity(t *testing.T) {
	_, err := LookupFunction("foobar")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Function "foobar" is not registered`)

	fn, err := LookupFunction("concat")
	require.NoError(t, err)
	require.Error(t, checkArgCount(fn, 1))
	require.NoError(t, checkArgCount(fn, 9))

	_, err = Eval(Fn("lower"), dmodel.DynamicFields{})
	require.Error(t, err)

	assert.Len(t, FunctionNames(), 16)
}
