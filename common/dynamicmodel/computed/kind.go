package computed

import (
	"go.bryk.io/pkg/errors"
)

// ComputeKind discriminates how a computed field's value is produced. This phase implements the
// two Go-executed kinds; aggregate/exists/lookup need SQL and belong to a later phase.
type ComputeKind string

const (
	// ComputeExpression evaluates an expression tree over the row's own fields.
	ComputeExpression ComputeKind = "expression"
	// ComputeRelated copies a scalar leaf reached through a chain of to-one edges.
	ComputeRelated ComputeKind = "related"
)

// Definition describes how a computed field gets its value. It is attached to a ModelField
// (stored there as an untyped value to avoid an import cycle; see DefOf) and is immutable after
// schema build.
type Definition struct {
	Kind ComputeKind

	// IsStored is false when the value is computed at read time (this phase). True — compute at
	// write time with source-change propagation — is reserved for a future phase and rejected by
	// validation, so the schema wire format never has to change when it arrives.
	IsStored bool

	// Expression is the evaluated tree when Kind is ComputeExpression.
	Expression Expr

	// Related is the dotted to-one edge path when Kind is ComputeRelated, e.g. "template.name".
	Related string
}

// NewDefinition builds a Definition from a root expression, deriving the kind: a RelatedExpr
// root means "related", anything else means "expression". This is what lets the field builder
// accept both forms through one parameter: Computed(false, computed.Sub(...)) and
// Computed(false, computed.Related("template.name")).
func NewDefinition(isStored bool, root Expr) (*Definition, error) {
	if root == nil {
		return nil, errors.New("computed definition requires an expression")
	}
	if related, ok := root.(RelatedExpr); ok {
		return &Definition{Kind: ComputeRelated, IsStored: isStored, Related: related.Path}, nil
	}
	if nested := findNestedRelated(root); nested != "" {
		return nil, errors.Errorf(
			"related reference %q is only allowed as the whole definition, not inside an expression", nested)
	}
	return &Definition{Kind: ComputeExpression, IsStored: isStored, Expression: root}, nil
}

// findNestedRelated returns the path of the first RelatedExpr found below the root, or "" when
// the tree is clean. A related copy costs a second query, so it must never hide inside arithmetic.
func findNestedRelated(expr Expr) string {
	switch node := expr.(type) {
	case RelatedExpr:
		return node.Path
	case BinaryExpr:
		if path := findNestedRelated(node.Left); path != "" {
			return path
		}
		return findNestedRelated(node.Right)
	case UnaryExpr:
		return findNestedRelated(node.Operand)
	case FunctionExpr:
		return firstNestedRelated(node.Args)
	case CaseExpr:
		return findNestedRelatedInCase(node)
	}
	return ""
}

func findNestedRelatedInCase(node CaseExpr) string {
	for _, branch := range node.Whens {
		if path := findNestedRelated(branch.When); path != "" {
			return path
		}
		if path := findNestedRelated(branch.Then); path != "" {
			return path
		}
	}
	if node.Else != nil {
		return findNestedRelated(node.Else)
	}
	return ""
}

func firstNestedRelated(exprs []Expr) string {
	for _, arg := range exprs {
		if path := findNestedRelated(arg); path != "" {
			return path
		}
	}
	return ""
}
