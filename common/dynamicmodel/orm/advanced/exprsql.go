package advanced

import (
	"fmt"
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// Go-evaluated computed kinds rendered to SQL.
//
// A related field is an alias for a physical column one to-one hop away, so it resolves to
// that column through the same join the caller could have written by hand. An expression field
// compiles its tree to a SQL expression over its (recursively resolved) operands. Both keep the
// Go evaluator's NULL semantics: every operator and builtin propagates NULL, comparisons on NULL
// are NULL, CASE treats a NULL condition as not matched, and division always happens in numeric.
//
// On the root's projection these kinds are still left to the Go evaluator (a service fills them
// after the read, and would overwrite a projected value anyway); the SQL rendering is what makes
// them filterable and sortable, and what projects them inside nested edge rows.

// internalPathDots bounds paths the builder writes for itself (a related rewrite appended to
// the requested prefix). The caller-facing caps apply to the requested path only.
const internalPathDots = MaxFilterDots + MaxSelectDots

func (this *queryPlan) resolveRelated(
	path string, owner *dmodel.ModelSchema, field *dmodel.ModelField, fieldPlan *computed.FieldPlan, u use,
) (*resolvedRef, error) {
	if u == useSelect && owner == this.root {
		return &resolvedRef{field: field, kind: refComputedGo, owner: owner, path: path}, nil
	}
	if fieldPlan.RelatedEdge == "" || fieldPlan.RelatedLeaf == "" {
		return nil, errors.Errorf("related field %q on %s has no resolved edge/leaf", field.Name(), owner.Name())
	}
	if err := this.rejectFanOutOwner(path); err != nil {
		return nil, err
	}
	target := joinPrefix(path) + fieldPlan.RelatedEdge + "." + fieldPlan.RelatedLeaf
	ref, err := this.resolveWith(target, u, internalPathDots)
	if err != nil {
		return nil, err
	}
	// The request asked for the related field's name, so that is the name the row carries.
	ref.path = path
	return ref, nil
}

func (this *queryPlan) resolveExpression(
	path string, owner *dmodel.ModelSchema, ownerAlias string, field *dmodel.ModelField,
	fieldPlan *computed.FieldPlan, u use,
) (*resolvedRef, error) {
	if u == useSelect && owner == this.root {
		return &resolvedRef{field: field, kind: refComputedGo, owner: owner, path: path}, nil
	}
	if err := this.rejectFanOutOwner(path); err != nil {
		return nil, err
	}
	if this.exprDepth >= computed.ActiveLimits().MaxComputedDependencyDepth {
		return nil, orm.WrapClientErrors(clientErrorsComputedNotFilterable(path,
			"computed expression nests deeper than the dependency limit"))
	}
	this.exprDepth++
	defer func() { this.exprDepth-- }()

	compiler := &exprCompiler{plan: this, prefix: joinPrefix(path), rootPath: path, u: u}
	sql, err := compiler.compile(fieldPlan.Def.Expression)
	if err != nil {
		return nil, err
	}
	return &resolvedRef{
		field: syntheticField(field), kind: refComputedExpr, owner: owner, path: path, sql: sql,
	}, nil
}

// joinPrefix is the edge chain of a path with its trailing dot, or "" for a root path.
func joinPrefix(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx < 0 {
		return ""
	}
	return path[:idx+1]
}

type exprCompiler struct {
	plan     *queryPlan
	prefix   string
	rootPath string
	u        use
}

func (this *exprCompiler) compile(expr computed.Expr) (string, error) {
	switch node := expr.(type) {
	case computed.FieldExpr:
		ref, err := this.plan.resolveWith(this.prefix+node.Name, this.u, internalPathDots)
		if err != nil {
			return "", err
		}
		if ref.kind == refComputedGo {
			return "", orm.WrapClientErrors(clientErrorsComputedNotFilterable(this.rootPath,
				fmt.Sprintf("operand %q is evaluated in Go and cannot be compiled to SQL", node.Name)))
		}
		return ref.sql, nil
	case computed.LiteralExpr:
		if node.Value == nil {
			return "NULL", nil
		}
		return orm.SqlLiteral(node.Value)
	case computed.BinaryExpr:
		return this.compileBinary(node)
	case computed.UnaryExpr:
		return this.compileUnary(node)
	case computed.FunctionExpr:
		return this.compileFunction(node)
	case computed.CaseExpr:
		return this.compileCase(node)
	}
	return "", errors.Errorf("computed expression: %T cannot compile to SQL", expr)
}

func (this *exprCompiler) compileBinary(node computed.BinaryExpr) (string, error) {
	left, err := this.compile(node.Left)
	if err != nil {
		return "", err
	}
	right, err := this.compile(node.Right)
	if err != nil {
		return "", err
	}
	operator, ok := sqlBinaryOperators[node.Op]
	if !ok {
		return "", errors.Errorf("computed expression: operator %q cannot compile to SQL", node.Op)
	}
	if node.Op == computed.OpDivide {
		// The Go evaluator divides in decimal; casting the dividend keeps SQL from integer-dividing.
		left = "(" + left + ")::numeric"
	}
	return "(" + left + " " + operator + " " + right + ")", nil
}

var sqlBinaryOperators = map[computed.BinaryOperator]string{
	computed.OpAdd:          "+",
	computed.OpSubtract:     "-",
	computed.OpMultiply:     "*",
	computed.OpDivide:       "/",
	computed.OpModulo:       "%",
	computed.OpEquals:       "=",
	computed.OpNotEquals:    "<>",
	computed.OpGreaterThan:  ">",
	computed.OpGreaterEqual: ">=",
	computed.OpLessThan:     "<",
	computed.OpLessEqual:    "<=",
	computed.OpAnd:          "AND",
	computed.OpOr:           "OR",
}

func (this *exprCompiler) compileUnary(node computed.UnaryExpr) (string, error) {
	operand, err := this.compile(node.Operand)
	if err != nil {
		return "", err
	}
	switch node.Op {
	case computed.OpNot:
		return "(NOT " + operand + ")", nil
	case computed.OpNegate:
		return "(- " + operand + ")", nil
	case computed.OpIsNull:
		return "(" + operand + " IS NULL)", nil
	case computed.OpIsNotNull:
		return "(" + operand + " IS NOT NULL)", nil
	}
	return "", errors.Errorf("computed expression: operator %q cannot compile to SQL", node.Op)
}

func (this *exprCompiler) compileFunction(node computed.FunctionExpr) (string, error) {
	args := make([]string, len(node.Args))
	for i, arg := range node.Args {
		compiled, err := this.compile(arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}
	joined := strings.Join(args, ", ")
	switch node.Name {
	case "coalesce":
		return "COALESCE(" + joined + ")", nil
	case "nullif":
		return "NULLIF(" + joined + ")", nil
	case "concat":
		// "||" propagates NULL like the Go builtin; CONCAT() would skip NULL operands instead.
		return "(" + strings.Join(args, " || ") + ")", nil
	case "lower":
		return "LOWER(" + joined + ")", nil
	case "upper":
		return "UPPER(" + joined + ")", nil
	case "trim":
		return "BTRIM(" + joined + ")", nil
	case "length":
		return "LENGTH(" + joined + ")", nil
	case "abs":
		return "ABS(" + joined + ")", nil
	case "ceil":
		return "CEIL(" + joined + ")", nil
	case "floor":
		return "FLOOR(" + joined + ")", nil
	case "round":
		if len(args) == 2 {
			return "ROUND((" + args[0] + ")::numeric, " + args[1] + ")", nil
		}
		return "ROUND(" + args[0] + ")", nil
	}
	// Date/time builtins have Go-only semantics (calendar arithmetic on the model wrappers);
	// compiling them would risk disagreeing with what the evaluator returns.
	return "", orm.WrapClientErrors(clientErrorsComputedNotFilterable(this.rootPath,
		fmt.Sprintf("function %q has no SQL rendering", node.Name)))
}

func (this *exprCompiler) compileCase(node computed.CaseExpr) (string, error) {
	var sb strings.Builder
	sb.WriteString("(CASE")
	for _, branch := range node.Whens {
		when, err := this.compile(branch.When)
		if err != nil {
			return "", err
		}
		then, err := this.compile(branch.Then)
		if err != nil {
			return "", err
		}
		sb.WriteString(" WHEN " + when + " THEN " + then)
	}
	fallback, err := this.compile(node.Else)
	if err != nil {
		return "", err
	}
	sb.WriteString(" ELSE " + fallback + " END)")
	return sb.String(), nil
}
