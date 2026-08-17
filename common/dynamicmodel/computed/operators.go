package computed

// BinaryOperator names a two-operand operation in a computed expression. The names are the wire
// values the JSON DSL uses, so they are full words rather than symbols: a schema author writes
// {"op": "multiply"}, never {"op": "*"}.
type BinaryOperator string

const (
	OpAdd      BinaryOperator = "add"
	OpSubtract BinaryOperator = "subtract"
	OpMultiply BinaryOperator = "multiply"
	OpDivide   BinaryOperator = "divide"
	OpModulo   BinaryOperator = "modulo"

	OpEquals       BinaryOperator = "eq"
	OpNotEquals    BinaryOperator = "neq"
	OpGreaterThan  BinaryOperator = "gt"
	OpGreaterEqual BinaryOperator = "gte"
	OpLessThan     BinaryOperator = "lt"
	OpLessEqual    BinaryOperator = "lte"

	OpAnd BinaryOperator = "and"
	OpOr  BinaryOperator = "or"
)

// IsArithmetic reports whether the operator combines two numeric operands into a numeric result.
func (op BinaryOperator) IsArithmetic() bool {
	switch op {
	case OpAdd, OpSubtract, OpMultiply, OpDivide, OpModulo:
		return true
	}
	return false
}

// IsComparison reports whether the operator compares two same-family operands into a boolean.
func (op BinaryOperator) IsComparison() bool {
	switch op {
	case OpEquals, OpNotEquals, OpGreaterThan, OpGreaterEqual, OpLessThan, OpLessEqual:
		return true
	}
	return false
}

// IsBoolean reports whether the operator combines two boolean operands into a boolean.
func (op BinaryOperator) IsBoolean() bool {
	return op == OpAnd || op == OpOr
}

// IsValid reports whether the operator is one of the supported binary operators.
func (op BinaryOperator) IsValid() bool {
	return op.IsArithmetic() || op.IsComparison() || op.IsBoolean()
}

// UnaryOperator names a one-operand operation in a computed expression.
type UnaryOperator string

const (
	OpNot       UnaryOperator = "not"
	OpNegate    UnaryOperator = "negate"
	OpIsNull    UnaryOperator = "is_null"
	OpIsNotNull UnaryOperator = "is_not_null"
)

// IsValid reports whether the operator is one of the supported unary operators.
func (op UnaryOperator) IsValid() bool {
	switch op {
	case OpNot, OpNegate, OpIsNull, OpIsNotNull:
		return true
	}
	return false
}
