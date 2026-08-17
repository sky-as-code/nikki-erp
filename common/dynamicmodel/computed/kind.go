package computed

import (
	"fmt"

	"go.bryk.io/pkg/errors"
)

// ComputeKind discriminates how a computed field's value is produced. Expression and related are
// Go-executed after the read; aggregate, exists and lookup compile to one correlated scalar
// subquery each, projected inside the SELECT.
type ComputeKind string

const (
	// ComputeExpression evaluates an expression tree over the row's own fields.
	ComputeExpression ComputeKind = "expression"
	// ComputeRelated copies a scalar leaf reached through a chain of to-one edges.
	ComputeRelated ComputeKind = "related"
	// ComputeAggregate aggregates a collection edge (COUNT/SUM/AVG/MIN/MAX) in SQL.
	ComputeAggregate ComputeKind = "aggregate"
	// ComputeExists is a boolean EXISTS(SELECT 1 ...) over a filtered collection edge.
	ComputeExists ComputeKind = "exists"
	// ComputeLookup copies one scalar from the first source record after filter + ordering.
	ComputeLookup ComputeKind = "lookup"
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

	// Exactly one of these is set for the SQL-compiled kinds.
	Aggregate *AggregateExpr
	Exists    *ExistsExpr
	Lookup    *LookupExpr
}

// NewDefinition builds a Definition from a root expression, deriving the kind from the root's
// type. This is what lets the field builder accept every kind through one parameter:
// Computed(false, computed.Sub(...)), Computed(false, computed.Related("template.name")),
// Computed(false, computed.Aggregate("lines", computed.AggCount)).
func NewDefinition(isStored bool, root Expr) (*Definition, error) {
	if root == nil {
		return nil, errors.New("computed definition requires an expression")
	}
	if def, matched, err := newSqlKindDefinition(isStored, root); matched {
		return def, err
	}
	if related, ok := root.(RelatedExpr); ok {
		return &Definition{Kind: ComputeRelated, IsStored: isStored, Related: related.Path}, nil
	}
	if nested := findRootOnlyNode(root); nested != "" {
		return nil, errors.Errorf(
			"%s is only allowed as the whole definition, not inside an expression", nested)
	}
	return &Definition{Kind: ComputeExpression, IsStored: isStored, Expression: root}, nil
}

// newSqlKindDefinition matches the SQL-compiled roots. The node's structural validation runs
// here so the chained API and the JSON parser share one chokepoint.
func newSqlKindDefinition(isStored bool, root Expr) (*Definition, bool, error) {
	var def *Definition
	var err error
	switch node := root.(type) {
	case AggregateExpr:
		def = &Definition{Kind: ComputeAggregate, IsStored: isStored, Aggregate: &node}
		err = node.validate()
	case ExistsExpr:
		def = &Definition{Kind: ComputeExists, IsStored: isStored, Exists: &node}
		err = node.validate()
	case LookupExpr:
		def = &Definition{Kind: ComputeLookup, IsStored: isStored, Lookup: &node}
		err = node.validate()
	default:
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	return def, true, nil
}

// findRootOnlyNode returns a description of the first root-only node found BELOW the root, or ""
// when the tree is clean. A related copy or a subquery costs a second query, so it must never
// hide inside arithmetic.
func findRootOnlyNode(expr Expr) string {
	switch node := expr.(type) {
	case RelatedExpr:
		return fmt.Sprintf("related reference %q", node.Path)
	case AggregateExpr:
		return fmt.Sprintf("aggregate over edge %q", node.Source)
	case ExistsExpr:
		return fmt.Sprintf("exists check over edge %q", node.Source)
	case LookupExpr:
		return fmt.Sprintf("lookup over edge %q", node.Source)
	case BinaryExpr:
		if found := findRootOnlyNode(node.Left); found != "" {
			return found
		}
		return findRootOnlyNode(node.Right)
	case UnaryExpr:
		return findRootOnlyNode(node.Operand)
	case FunctionExpr:
		return firstRootOnlyNode(node.Args)
	case CaseExpr:
		return findRootOnlyNodeInCase(node)
	}
	return ""
}

func findRootOnlyNodeInCase(node CaseExpr) string {
	for _, branch := range node.Whens {
		if found := findRootOnlyNode(branch.When); found != "" {
			return found
		}
		if found := findRootOnlyNode(branch.Then); found != "" {
			return found
		}
	}
	if node.Else != nil {
		return findRootOnlyNode(node.Else)
	}
	return ""
}

func firstRootOnlyNode(exprs []Expr) string {
	for _, arg := range exprs {
		if found := findRootOnlyNode(arg); found != "" {
			return found
		}
	}
	return ""
}
