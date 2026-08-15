package computed

import (
	"strings"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Eval computes an expression's value from one row. It preserves SQL NULL semantics (spec §33):
// arithmetic and comparison over nil yield nil, boolean operators use three-valued logic, and
// nothing is ever implicitly converted to zero, empty string or false. Type CORRECTNESS is the
// schema validator's job — by the time Eval runs, the tree has passed inference, so a coercion
// failure here is a server fault, not a client error.
func Eval(expr Expr, row dmodel.DynamicFields) (any, error) {
	switch node := expr.(type) {
	case FieldExpr:
		return normalizeValue(row[node.Name]), nil
	case LiteralExpr:
		return normalizeValue(node.Value), nil
	case BinaryExpr:
		return evalBinary(node, row)
	case UnaryExpr:
		return evalUnary(node, row)
	case FunctionExpr:
		return evalFunction(node, row)
	case CaseExpr:
		return evalCase(node, row)
	case RelatedExpr:
		return nil, errors.New("related expression is resolved by the eval plan, not the row evaluator")
	}
	return nil, errors.Errorf("unsupported expression node %T", expr)
}

func evalBinary(node BinaryExpr, row dmodel.DynamicFields) (any, error) {
	left, err := Eval(node.Left, row)
	if err != nil {
		return nil, err
	}
	right, err := Eval(node.Right, row)
	if err != nil {
		return nil, err
	}
	switch {
	case node.Op.IsArithmetic():
		return evalArithmetic(node.Op, left, right)
	case node.Op.IsComparison():
		return evalComparison(node.Op, left, right)
	case node.Op.IsBoolean():
		return evalThreeValued(node.Op, left, right)
	}
	return nil, errors.Errorf("unsupported binary operator %q", node.Op)
}

func evalArithmetic(op BinaryOperator, left any, right any) (any, error) {
	if left == nil || right == nil {
		return nil, nil
	}
	a, err := coerceDecimal(left)
	if err != nil {
		return nil, err
	}
	b, err := coerceDecimal(right)
	if err != nil {
		return nil, err
	}
	result, err := applyArithmetic(op, a, b)
	if err != nil {
		return nil, err
	}
	// Integer op integer stays integer, except division which always widens to decimal.
	if op != OpDivide && isIntegerValue(left) && isIntegerValue(right) {
		return result.IntPart(), nil
	}
	return result, nil
}

func applyArithmetic(op BinaryOperator, a decimal.Decimal, b decimal.Decimal) (decimal.Decimal, error) {
	switch op {
	case OpAdd:
		return a.Add(b), nil
	case OpSubtract:
		return a.Sub(b), nil
	case OpMultiply:
		return a.Mul(b), nil
	case OpDivide:
		if b.IsZero() {
			return decimal.Zero, errors.New("division by zero in computed expression")
		}
		return a.Div(b), nil
	case OpModulo:
		if b.IsZero() {
			return decimal.Zero, errors.New("division by zero in computed expression")
		}
		return a.Mod(b), nil
	}
	return decimal.Zero, errors.Errorf("operator %q is not arithmetic", op)
}

func evalComparison(op BinaryOperator, left any, right any) (any, error) {
	if left == nil || right == nil {
		return nil, nil
	}
	ordering, err := compareOrder(left, right)
	if err != nil {
		return nil, err
	}
	switch op {
	case OpEquals:
		return ordering == 0, nil
	case OpNotEquals:
		return ordering != 0, nil
	case OpGreaterThan:
		return ordering > 0, nil
	case OpGreaterEqual:
		return ordering >= 0, nil
	case OpLessThan:
		return ordering < 0, nil
	case OpLessEqual:
		return ordering <= 0, nil
	}
	return nil, errors.Errorf("operator %q is not a comparison", op)
}

// compareOrder orders two same-family values: negative, zero or positive.
func compareOrder(left any, right any) (int, error) {
	if isIntegerValue(left) || isDecimalValue(left) {
		return compareNumeric(left, right)
	}
	switch lv := left.(type) {
	case string:
		rv, err := coerceString(right)
		if err != nil {
			return 0, err
		}
		return strings.Compare(lv, rv), nil
	case bool:
		return compareBool(lv, right)
	}
	return compareTemporal(left, right)
}

func compareNumeric(left any, right any) (int, error) {
	a, err := coerceDecimal(left)
	if err != nil {
		return 0, err
	}
	b, err := coerceDecimal(right)
	if err != nil {
		return 0, err
	}
	return a.Cmp(b), nil
}

func compareBool(left bool, right any) (int, error) {
	rv, err := coerceBool(right)
	if err != nil {
		return 0, err
	}
	if left == rv {
		return 0, nil
	}
	if !left {
		return -1, nil
	}
	return 1, nil
}

func compareTemporal(left any, right any) (int, error) {
	a, err := coerceTime(left)
	if err != nil {
		return 0, err
	}
	b, err := coerceTime(right)
	if err != nil {
		return 0, err
	}
	return a.Compare(b), nil
}

// compareEquality reports whether two non-nil values are equal, used by nullif.
func compareEquality(left any, right any) (bool, error) {
	ordering, err := compareOrder(left, right)
	if err != nil {
		return false, err
	}
	return ordering == 0, nil
}

func isDecimalValue(value any) bool {
	_, ok := value.(decimal.Decimal)
	return ok
}

// evalThreeValued implements SQL three-valued AND/OR: false AND null is false, true OR null is
// true, and only the undecidable combinations yield nil.
func evalThreeValued(op BinaryOperator, left any, right any) (any, error) {
	a, err := toThreeValued(left)
	if err != nil {
		return nil, err
	}
	b, err := toThreeValued(right)
	if err != nil {
		return nil, err
	}
	if op == OpAnd {
		return threeValuedAnd(a, b), nil
	}
	return threeValuedOr(a, b), nil
}

func toThreeValued(value any) (*bool, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := coerceBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func threeValuedAnd(a *bool, b *bool) any {
	if (a != nil && !*a) || (b != nil && !*b) {
		return false
	}
	if a == nil || b == nil {
		return nil
	}
	return true
}

func threeValuedOr(a *bool, b *bool) any {
	if (a != nil && *a) || (b != nil && *b) {
		return true
	}
	if a == nil || b == nil {
		return nil
	}
	return false
}

func evalUnary(node UnaryExpr, row dmodel.DynamicFields) (any, error) {
	operand, err := Eval(node.Operand, row)
	if err != nil {
		return nil, err
	}
	switch node.Op {
	case OpIsNull:
		return operand == nil, nil
	case OpIsNotNull:
		return operand != nil, nil
	case OpNot:
		return evalNot(operand)
	case OpNegate:
		return evalNegate(operand)
	}
	return nil, errors.Errorf("unsupported unary operator %q", node.Op)
}

func evalNot(operand any) (any, error) {
	if operand == nil {
		return nil, nil
	}
	value, err := coerceBool(operand)
	if err != nil {
		return nil, err
	}
	return !value, nil
}

func evalNegate(operand any) (any, error) {
	if operand == nil {
		return nil, nil
	}
	value, err := coerceDecimal(operand)
	if err != nil {
		return nil, err
	}
	if isIntegerValue(operand) {
		return value.Neg().IntPart(), nil
	}
	return value.Neg(), nil
}

func evalFunction(node FunctionExpr, row dmodel.DynamicFields) (any, error) {
	fn, err := LookupFunction(node.Name)
	if err != nil {
		return nil, err
	}
	if err := checkArgCount(fn, len(node.Args)); err != nil {
		return nil, err
	}
	args := make([]any, len(node.Args))
	for i, argExpr := range node.Args {
		arg, err := Eval(argExpr, row)
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}
	return fn.Eval(args)
}

// evalCase picks the first branch whose condition is true. A nil condition counts as not matched,
// exactly like SQL CASE.
func evalCase(node CaseExpr, row dmodel.DynamicFields) (any, error) {
	for _, branch := range node.Whens {
		cond, err := Eval(branch.When, row)
		if err != nil {
			return nil, err
		}
		matched, err := toThreeValued(cond)
		if err != nil {
			return nil, err
		}
		if matched != nil && *matched {
			return Eval(branch.Then, row)
		}
	}
	return Eval(node.Else, row)
}
