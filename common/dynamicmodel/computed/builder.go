package computed

// Chained constructors for authoring computed expressions in Go. They build pure AST values with
// no registry lookups and no side effects, so a definition can be declared at package level and
// validated later, at schema finalize time:
//
//	Computed(false, computed.Sub(computed.F("on_hand_quantity"), computed.F("reserved_quantity")))
//	Computed(false, computed.Related("template.name"))

// F references a same-row field by name.
func F(name string) Expr {
	return FieldExpr{Name: name}
}

// Lit embeds a constant operand.
func Lit(value any) Expr {
	return LiteralExpr{Value: value}
}

// Related copies a scalar leaf reached through a chain of to-one edges, e.g. "template.name".
// Only legal as the whole definition — see RelatedExpr.
func Related(path string) Expr {
	return RelatedExpr{Path: path}
}

func Add(left Expr, right Expr) Expr    { return BinaryExpr{Op: OpAdd, Left: left, Right: right} }
func Sub(left Expr, right Expr) Expr    { return BinaryExpr{Op: OpSubtract, Left: left, Right: right} }
func Mul(left Expr, right Expr) Expr    { return BinaryExpr{Op: OpMultiply, Left: left, Right: right} }
func Div(left Expr, right Expr) Expr    { return BinaryExpr{Op: OpDivide, Left: left, Right: right} }
func Mod(left Expr, right Expr) Expr    { return BinaryExpr{Op: OpModulo, Left: left, Right: right} }
func Eq(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpEquals, Left: left, Right: right} }
func Ne(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpNotEquals, Left: left, Right: right} }
func Gt(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpGreaterThan, Left: left, Right: right} }
func Ge(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpGreaterEqual, Left: left, Right: right} }
func Lt(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpLessThan, Left: left, Right: right} }
func Le(left Expr, right Expr) Expr     { return BinaryExpr{Op: OpLessEqual, Left: left, Right: right} }
func Not(operand Expr) Expr             { return UnaryExpr{Op: OpNot, Operand: operand} }
func Neg(operand Expr) Expr             { return UnaryExpr{Op: OpNegate, Operand: operand} }
func IsNull(operand Expr) Expr          { return UnaryExpr{Op: OpIsNull, Operand: operand} }
func IsNotNull(operand Expr) Expr       { return UnaryExpr{Op: OpIsNotNull, Operand: operand} }
func Fn(name string, args ...Expr) Expr { return FunctionExpr{Name: name, Args: args} }

// And folds two or more boolean operands left-to-right.
func And(first Expr, second Expr, rest ...Expr) Expr {
	return foldBinary(OpAnd, first, second, rest)
}

// Or folds two or more boolean operands left-to-right.
func Or(first Expr, second Expr, rest ...Expr) Expr {
	return foldBinary(OpOr, first, second, rest)
}

func foldBinary(op BinaryOperator, first Expr, second Expr, rest []Expr) Expr {
	acc := BinaryExpr{Op: op, Left: first, Right: second}
	for _, next := range rest {
		acc = BinaryExpr{Op: op, Left: acc, Right: next}
	}
	return acc
}

// CaseBuilder accumulates WHEN/THEN branches; Else finishes the expression. Else is required by
// the AST, so the builder has no way to produce a CaseExpr without it.
type CaseBuilder struct {
	whens []WhenThen
}

// Case starts a conditional expression: Case().When(cond, then).Else(fallback).
func Case() *CaseBuilder {
	return &CaseBuilder{}
}

// When appends a branch. Cond must evaluate to boolean; validation enforces it.
func (b *CaseBuilder) When(cond Expr, then Expr) *CaseBuilder {
	b.whens = append(b.whens, WhenThen{When: cond, Then: then})
	return b
}

// Else completes the CaseExpr with the value used when no branch matches.
func (b *CaseBuilder) Else(fallback Expr) Expr {
	return CaseExpr{Whens: b.whens, Else: fallback}
}
