package computed

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"
)

// FieldTypeResolver supplies the type of a referenced field. The schema-side caller backs it
// with the registry, so an unknown field name surfaces here as an error during validation.
type FieldTypeResolver func(fieldName string) (Type, error)

// InferType statically types an expression tree. It runs at schema validation time and rejects
// invalid operand combinations (string * date, non-boolean AND, un-unifiable CASE branches), so
// evaluation never has to make a typing decision per row.
func InferType(expr Expr, resolve FieldTypeResolver) (Type, error) {
	switch node := expr.(type) {
	case FieldExpr:
		return resolve(node.Name)
	case LiteralExpr:
		return literalType(node.Value)
	case BinaryExpr:
		return inferBinary(node, resolve)
	case UnaryExpr:
		return inferUnary(node, resolve)
	case FunctionExpr:
		return inferFunction(node, resolve)
	case CaseExpr:
		return inferCase(node, resolve)
	case RelatedExpr:
		return TypeUnknown, errors.New("related reference is typed by the schema resolver, not expression inference")
	}
	return TypeUnknown, errors.Errorf("unsupported expression node %T", expr)
}

func literalType(value any) (Type, error) {
	normalized := normalizeValue(value)
	if normalized == nil {
		return TypeNull, nil
	}
	switch normalized.(type) {
	case string:
		return TypeString, nil
	case bool:
		return TypeBoolean, nil
	case int, int32:
		return TypeInt32, nil
	case int64:
		return TypeInt64, nil
	case float32, float64, decimal.Decimal:
		return TypeDecimal, nil
	case time.Time:
		return TypeDateTime, nil
	}
	return TypeUnknown, errors.Errorf("literal %v (%T) has no supported computed type", value, value)
}

func inferBinary(node BinaryExpr, resolve FieldTypeResolver) (Type, error) {
	left, err := InferType(node.Left, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	right, err := InferType(node.Right, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	switch {
	case node.Op.IsArithmetic():
		return inferArithmetic(node.Op, left, right)
	case node.Op.IsComparison():
		return inferComparison(node.Op, left, right)
	case node.Op.IsBoolean():
		return inferBooleanOp(node.Op, left, right)
	}
	return TypeUnknown, errors.Errorf("operator %q is not a supported binary operator", node.Op)
}

func inferArithmetic(op BinaryOperator, left Type, right Type) (Type, error) {
	if (left != TypeNull && !left.IsNumeric()) || (right != TypeNull && !right.IsNumeric()) {
		return TypeUnknown, errors.Errorf(
			"operator %q expects numeric operands but received %s and %s", op, orNullName(left), orNullName(right))
	}
	if op == OpDivide {
		return TypeDecimal, nil
	}
	// A null operand takes the other side's type: the result is null at evaluation time, but the
	// expression still has the non-null operand's static type, as in SQL.
	if left == TypeNull {
		return numericPassthroughType(right), nil
	}
	if right == TypeNull {
		return numericPassthroughType(left), nil
	}
	return widenNumeric(left, right), nil
}

func inferComparison(op BinaryOperator, left Type, right Type) (Type, error) {
	if !left.ComparableWith(right) {
		return TypeUnknown, errors.Errorf(
			"operator %q cannot compare %s with %s", op, orNullName(left), orNullName(right))
	}
	return TypeBoolean, nil
}

func inferBooleanOp(op BinaryOperator, left Type, right Type) (Type, error) {
	if (left != TypeNull && left != TypeBoolean) || (right != TypeNull && right != TypeBoolean) {
		return TypeUnknown, errors.Errorf(
			"operator %q expects boolean operands but received %s and %s", op, orNullName(left), orNullName(right))
	}
	return TypeBoolean, nil
}

func inferUnary(node UnaryExpr, resolve FieldTypeResolver) (Type, error) {
	operand, err := InferType(node.Operand, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	switch node.Op {
	case OpIsNull, OpIsNotNull:
		return TypeBoolean, nil
	case OpNot:
		if operand != TypeNull && operand != TypeBoolean {
			return TypeUnknown, errors.Errorf("operator \"not\" expects a boolean operand but received %s", operand)
		}
		return TypeBoolean, nil
	case OpNegate:
		if operand != TypeNull && !operand.IsNumeric() {
			return TypeUnknown, errors.Errorf("operator \"negate\" expects a numeric operand but received %s", operand)
		}
		return numericPassthroughType(operand), nil
	}
	return TypeUnknown, errors.Errorf("operator %q is not a supported unary operator", node.Op)
}

func inferFunction(node FunctionExpr, resolve FieldTypeResolver) (Type, error) {
	fn, err := LookupFunction(node.Name)
	if err != nil {
		return TypeUnknown, err
	}
	if err := checkArgCount(fn, len(node.Args)); err != nil {
		return TypeUnknown, err
	}
	args := make([]Type, len(node.Args))
	for i, argExpr := range node.Args {
		argType, err := InferType(argExpr, resolve)
		if err != nil {
			return TypeUnknown, err
		}
		args[i] = argType
	}
	return fn.ReturnType(args)
}

func inferCase(node CaseExpr, resolve FieldTypeResolver) (Type, error) {
	if len(node.Whens) == 0 || node.Else == nil {
		return TypeUnknown, errors.New("case expression requires at least one when branch and an else")
	}
	result, err := InferType(node.Else, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	for _, branch := range node.Whens {
		if result, err = inferCaseBranch(branch, result, resolve); err != nil {
			return TypeUnknown, err
		}
	}
	if result == TypeNull {
		return TypeUnknown, errors.New("case expression requires at least one non-null branch type")
	}
	return result, nil
}

func inferCaseBranch(branch WhenThen, unified Type, resolve FieldTypeResolver) (Type, error) {
	cond, err := InferType(branch.When, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	if cond != TypeNull && cond != TypeBoolean {
		return TypeUnknown, errors.Errorf("case condition must be boolean but received %s", cond)
	}
	then, err := InferType(branch.Then, resolve)
	if err != nil {
		return TypeUnknown, err
	}
	result, ok := unify(unified, then)
	if !ok {
		return TypeUnknown, errors.Errorf("case branches have incompatible types %s and %s", unified, then)
	}
	return result, nil
}

// orNullName renders TypeNull readably in error messages.
func orNullName(t Type) string {
	if t == TypeNull {
		return "null"
	}
	return string(t)
}
