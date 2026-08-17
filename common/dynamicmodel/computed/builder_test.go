package computed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_ArithmeticTreeShape(t *testing.T) {
	expr := Sub(F("on_hand_quantity"), F("reserved_quantity"))

	require.IsType(t, BinaryExpr{}, expr)
	bin := expr.(BinaryExpr)
	assert.Equal(t, OpSubtract, bin.Op)
	assert.Equal(t, FieldExpr{Name: "on_hand_quantity"}, bin.Left)
	assert.Equal(t, FieldExpr{Name: "reserved_quantity"}, bin.Right)
}

func TestBuilder_NestedExpressionEqualsHandBuiltTree(t *testing.T) {
	chained := Mul(Add(F("a"), Lit(2)), F("b"))
	handBuilt := BinaryExpr{
		Op: OpMultiply,
		Left: BinaryExpr{
			Op:    OpAdd,
			Left:  FieldExpr{Name: "a"},
			Right: LiteralExpr{Value: 2},
		},
		Right: FieldExpr{Name: "b"},
	}

	assert.Equal(t, handBuilt, chained)
}

func TestBuilder_AndOrFoldLeftToRight(t *testing.T) {
	expr := And(F("a"), F("b"), F("c"))

	bin := expr.(BinaryExpr)
	assert.Equal(t, OpAnd, bin.Op)
	assert.Equal(t, FieldExpr{Name: "c"}, bin.Right)

	inner := bin.Left.(BinaryExpr)
	assert.Equal(t, OpAnd, inner.Op)
	assert.Equal(t, FieldExpr{Name: "a"}, inner.Left)
	assert.Equal(t, FieldExpr{Name: "b"}, inner.Right)
}

func TestBuilder_FunctionAndUnary(t *testing.T) {
	expr := Fn("concat", F("code"), Lit(" - "), F("name"))

	fn := expr.(FunctionExpr)
	assert.Equal(t, "concat", fn.Name)
	require.Len(t, fn.Args, 3)
	assert.Equal(t, LiteralExpr{Value: " - "}, fn.Args[1])

	assert.Equal(t, UnaryExpr{Op: OpIsNull, Operand: FieldExpr{Name: "x"}}, IsNull(F("x")))
	assert.Equal(t, UnaryExpr{Op: OpNot, Operand: FieldExpr{Name: "x"}}, Not(F("x")))
}

func TestBuilder_CaseRequiresElse(t *testing.T) {
	expr := Case().
		When(Le(F("available_qty"), Lit(0)), Lit("out_of_stock")).
		When(Lt(F("available_qty"), F("reorder_point")), Lit("low_stock")).
		Else(Lit("in_stock"))

	caseExpr := expr.(CaseExpr)
	require.Len(t, caseExpr.Whens, 2)
	assert.Equal(t, LiteralExpr{Value: "in_stock"}, caseExpr.Else)
	assert.Equal(t, OpLessEqual, caseExpr.Whens[0].When.(BinaryExpr).Op)
}

func TestNewDefinition_DerivesExpressionKind(t *testing.T) {
	def, err := NewDefinition(false, Sub(F("on_hand_quantity"), F("reserved_quantity")))

	require.NoError(t, err)
	assert.Equal(t, ComputeExpression, def.Kind)
	assert.False(t, def.IsStored)
	assert.NotNil(t, def.Expression)
	assert.Empty(t, def.Related)
}

func TestNewDefinition_DerivesRelatedKind(t *testing.T) {
	def, err := NewDefinition(false, Related("template.name"))

	require.NoError(t, err)
	assert.Equal(t, ComputeRelated, def.Kind)
	assert.Equal(t, "template.name", def.Related)
	assert.Nil(t, def.Expression)
}

func TestNewDefinition_KeepsIsStoredFlag(t *testing.T) {
	def, err := NewDefinition(true, Related("template.name"))

	require.NoError(t, err)
	assert.True(t, def.IsStored)
}

func TestNewDefinition_NilExpressionRejected(t *testing.T) {
	def, err := NewDefinition(false, nil)

	require.Error(t, err)
	assert.Nil(t, def)
}

func TestNewDefinition_NestedRelatedRejected(t *testing.T) {
	cases := map[string]Expr{
		"inside arithmetic": Mul(Related("template.price"), Lit(2)),
		"inside function":   Fn("lower", Related("template.name")),
		"inside unary":      IsNull(Related("template.name")),
		"inside case then": Case().
			When(F("flag"), Related("template.name")).
			Else(Lit("")),
		"inside case else": Case().
			When(F("flag"), Lit("")).
			Else(Related("template.name")),
	}

	for name, expr := range cases {
		t.Run(name, func(t *testing.T) {
			def, err := NewDefinition(false, expr)
			require.Error(t, err)
			assert.Nil(t, def)
			assert.Contains(t, err.Error(), "only allowed as the whole definition")
		})
	}
}

func TestOperators_ClassificationIsDisjoint(t *testing.T) {
	arithmetic := []BinaryOperator{OpAdd, OpSubtract, OpMultiply, OpDivide, OpModulo}
	comparison := []BinaryOperator{OpEquals, OpNotEquals, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual}
	boolean := []BinaryOperator{OpAnd, OpOr}

	for _, op := range arithmetic {
		assert.True(t, op.IsArithmetic() && op.IsValid(), op)
		assert.False(t, op.IsComparison() || op.IsBoolean(), op)
	}
	for _, op := range comparison {
		assert.True(t, op.IsComparison() && op.IsValid(), op)
		assert.False(t, op.IsArithmetic() || op.IsBoolean(), op)
	}
	for _, op := range boolean {
		assert.True(t, op.IsBoolean() && op.IsValid(), op)
	}
	assert.False(t, BinaryOperator("bogus").IsValid())
	assert.False(t, UnaryOperator("bogus").IsValid())
	assert.True(t, UnaryOperator("is_not_null").IsValid())
}
