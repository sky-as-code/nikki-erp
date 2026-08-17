package computed

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolverFor(fields map[string]Type) FieldTypeResolver {
	return func(name string) (Type, error) {
		t, ok := fields[name]
		if !ok {
			return TypeUnknown, assert.AnError
		}
		return t, nil
	}
}

func TestInferType_ArithmeticWidening(t *testing.T) {
	resolve := resolverFor(map[string]Type{
		"i32": TypeInt32, "i64": TypeInt64, "dec": TypeDecimal,
	})

	cases := []struct {
		name string
		expr Expr
		want Type
	}{
		{"int32+int32", Add(F("i32"), F("i32")), TypeInt32},
		{"int32+int64", Add(F("i32"), F("i64")), TypeInt64},
		{"int64*decimal", Mul(F("i64"), F("dec")), TypeDecimal},
		{"decimal-decimal", Sub(F("dec"), F("dec")), TypeDecimal},
		{"division always decimal", Div(F("i32"), F("i32")), TypeDecimal},
		{"modulo keeps ints", Mod(F("i64"), F("i32")), TypeInt64},
		{"null yields other side", Add(F("i32"), Lit(nil)), TypeInt32},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InferType(tc.expr, resolve)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInferType_InvalidOperandsRejected(t *testing.T) {
	resolve := resolverFor(map[string]Type{
		"name": TypeString, "due": TypeDate, "qty": TypeDecimal, "ok": TypeBoolean,
	})

	cases := []struct {
		name string
		expr Expr
	}{
		{"string times date", Mul(F("name"), F("due"))},
		{"string plus number", Add(F("name"), F("qty"))},
		{"cross-family comparison", Gt(F("name"), F("qty"))},
		{"and over numbers", And(F("qty"), F("qty"))},
		{"not over string", Not(F("name"))},
		{"negate string", Neg(F("name"))},
		{"unknown field", Add(F("missing"), F("qty"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := InferType(tc.expr, resolve)
			require.Error(t, err)
		})
	}
}

func TestInferType_ComparisonAndBooleanShapes(t *testing.T) {
	resolve := resolverFor(map[string]Type{
		"qty": TypeDecimal, "min": TypeInt32, "ok": TypeBoolean, "due": TypeDate, "at": TypeDateTime,
	})

	for name, expr := range map[string]Expr{
		"numeric cross-width compare": Ge(F("qty"), F("min")),
		"temporal compare":            Lt(F("due"), F("at")),
		"boolean and":                 And(F("ok"), F("ok")),
		"is null of anything":         IsNull(F("due")),
		"comparison with null":        Eq(F("qty"), Lit(nil)),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := InferType(expr, resolve)
			require.NoError(t, err)
			assert.Equal(t, TypeBoolean, got)
		})
	}
}

func TestInferType_CaseBranchUnification(t *testing.T) {
	resolve := resolverFor(map[string]Type{
		"qty": TypeInt32, "flag": TypeBoolean, "label": TypeString,
	})

	t.Run("numeric branches widen", func(t *testing.T) {
		expr := Case().When(F("flag"), F("qty")).Else(Lit(decimal.NewFromInt(2)))
		got, err := InferType(expr, resolve)
		require.NoError(t, err)
		assert.Equal(t, TypeDecimal, got)
	})

	t.Run("string and int do not unify", func(t *testing.T) {
		expr := Case().When(F("flag"), F("label")).Else(F("qty"))
		_, err := InferType(expr, resolve)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible types")
	})

	t.Run("condition must be boolean", func(t *testing.T) {
		expr := Case().When(F("qty"), F("label")).Else(F("label"))
		_, err := InferType(expr, resolve)
		require.Error(t, err)
	})

	t.Run("null branch yields to the other", func(t *testing.T) {
		expr := Case().When(F("flag"), Lit(nil)).Else(F("label"))
		got, err := InferType(expr, resolve)
		require.NoError(t, err)
		assert.Equal(t, TypeString, got)
	})
}

func TestInferType_FunctionSignatures(t *testing.T) {
	resolve := resolverFor(map[string]Type{
		"code": TypeString, "qty": TypeDecimal, "n": TypeInt32, "due": TypeDate,
	})

	cases := []struct {
		name    string
		expr    Expr
		want    Type
		wantErr bool
	}{
		{"concat returns string", Fn("concat", F("code"), Lit("-")), TypeString, false},
		{"concat rejects numeric", Fn("concat", F("code"), F("qty")), TypeUnknown, true},
		{"length returns int32", Fn("length", F("code")), TypeInt32, false},
		{"round keeps family", Fn("round", F("qty"), F("n")), TypeDecimal, false},
		{"round of int stays int", Fn("round", F("n")), TypeInt32, false},
		{"abs rejects string", Fn("abs", F("code")), TypeUnknown, true},
		{"coalesce unifies", Fn("coalesce", F("qty"), Lit(0)), TypeDecimal, false},
		{"coalesce rejects mixed", Fn("coalesce", F("qty"), F("code")), TypeUnknown, true},
		{"today returns date", Fn("today"), TypeDate, false},
		{"now returns datetime", Fn("now"), TypeDateTime, false},
		{"date_add passthrough", Fn("date_add", F("due"), F("n")), TypeDate, false},
		{"date_diff returns int32", Fn("date_diff", F("due"), F("due")), TypeInt32, false},
		{"extract returns int32", Fn("extract", Lit("year"), F("due")), TypeInt32, false},
		{"unknown function", Fn("foobar", F("code")), TypeUnknown, true},
		{"wrong arity", Fn("lower", F("code"), F("code")), TypeUnknown, true},
		{"is overdue comparison", Lt(F("due"), Fn("today")), TypeBoolean, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InferType(tc.expr, resolve)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInferType_UnknownFunctionErrorWording(t *testing.T) {
	_, err := InferType(Fn("foobar"), resolverFor(nil))

	require.Error(t, err)
	assert.Contains(t, err.Error(), `Function "foobar" is not registered for computed expressions`)
}
