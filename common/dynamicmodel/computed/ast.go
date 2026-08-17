package computed

// Expr is a node in a computed field's expression tree. The tree is pure data: it is built by the
// chained constructors in builder.go or parsed from the JSON DSL, validated against the schema
// registry at finalize time, and evaluated in Go per row at read time. It never compiles to SQL
// in this phase — a not-stored computed field lives entirely outside the database.
type Expr interface {
	exprNode()
}

// FieldExpr references another field on the same schema by name. In the current phase the name
// must be a same-row field (physical or computed); validation resolves it through the schema
// registry, so a name that is not registered never survives to evaluation.
type FieldExpr struct {
	Name string
}

// LiteralExpr is a constant embedded in the expression. The value never reaches SQL: it is used
// only as a Go operand during evaluation.
type LiteralExpr struct {
	Value any
}

// BinaryExpr combines two operands with an arithmetic, comparison or boolean operator.
type BinaryExpr struct {
	Op    BinaryOperator
	Left  Expr
	Right Expr
}

// UnaryExpr applies a one-operand operator: not, negate, is_null, is_not_null.
type UnaryExpr struct {
	Op      UnaryOperator
	Operand Expr
}

// FunctionExpr calls a function from the registry by name. The name is looked up in the fixed,
// compile-time registry; an unregistered name is a schema validation error, so evaluation only
// ever runs functions baked into the binary.
type FunctionExpr struct {
	Name string
	Args []Expr
}

// WhenThen is one branch of a CaseExpr: When must evaluate to boolean, Then supplies the value.
type WhenThen struct {
	When Expr
	Then Expr
}

// CaseExpr picks the first branch whose condition holds, else the Else value. Else is required:
// branch-type unification stays unambiguous only when every path yields a value.
type CaseExpr struct {
	Whens []WhenThen
	Else  Expr
}

// RelatedExpr copies a scalar leaf reached through a chain of to-one edges, e.g. "template.name".
// It is only legal as the ROOT of a definition (kind "related"): a copy is a whole-field aliasing
// operation, not an operand, so nesting it inside arithmetic would hide a second query behind an
// innocent-looking expression. NewDefinition and validation both enforce root-only placement.
type RelatedExpr struct {
	Path string
}

func (FieldExpr) exprNode()    {}
func (LiteralExpr) exprNode()  {}
func (BinaryExpr) exprNode()   {}
func (UnaryExpr) exprNode()    {}
func (FunctionExpr) exprNode() {}
func (CaseExpr) exprNode()     {}
func (RelatedExpr) exprNode()  {}
