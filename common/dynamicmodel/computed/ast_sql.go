package computed

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The SQL-compiled computed kinds: aggregate, exists and lookup. Unlike the Go-evaluated
// expression tree, these nodes compile to one correlated scalar subquery each, projected inside
// the SELECT — so the row already carries the value when it reaches Go. Like RelatedExpr they are
// only legal as the ROOT of a definition: a subquery is a whole-field operation, and hiding one
// inside arithmetic would smuggle a query behind an innocent-looking expression.

// OrderBy is one ordering entry of a LookupExpr, over a field of the source schema.
type OrderBy struct {
	Field string
	Desc  bool
}

// AggregateExpr aggregates a collection edge of the schema: COUNT/SUM/AVG/MIN/MAX over either a
// single source field or a restricted inner expression (arithmetic over source fields — see the
// SQL-compilable subset in validation). Filter narrows the aggregated rows and reuses the
// existing search-node shapes; it never introduces a second filter language.
type AggregateExpr struct {
	Source   string
	Function AggregateFunction
	// Field is the aggregated source-schema column. Mutually exclusive with Expr; count uses
	// neither, count_distinct requires Field.
	Field string
	// Expr is the aggregated inner expression for sum/avg/min/max, e.g. quantity * unit_price.
	Expr   Expr
	Filter *dmodel.SearchNode
	// Context lists the whitelisted request-context keys the filter may reference as
	// "${ctx.key}" condition values.
	Context []string
	// Default replaces a NULL aggregate result (sum/avg/min/max over zero rows).
	Default any
}

// ExistsExpr is a boolean computed from the existence of source records matching the filter,
// compiled to EXISTS (SELECT 1 ...). No rows simply means false, so it takes no default.
type ExistsExpr struct {
	Source  string
	Filter  *dmodel.SearchNode
	Context []string
}

// LookupExpr copies one scalar from the first source record after filter + ordering, compiled
// with LIMIT 1. No matching record yields NULL unless Default is declared.
type LookupExpr struct {
	Source  string
	Field   string
	OrderBy []OrderBy
	Filter  *dmodel.SearchNode
	Context []string
	Default any
}

func (AggregateExpr) exprNode() {}
func (ExistsExpr) exprNode()    {}
func (LookupExpr) exprNode()    {}

// validate applies the structural rules both authoring forms share. It runs in the JSON parser
// (early feedback) and again in NewDefinition (the chokepoint the chained API goes through), so
// neither form can skip it. Registry-dependent checks (edge exists, fields exist) belong to
// finalize-time resolution, not here.
func (this AggregateExpr) validate() error {
	if err := requireSourceEdge("aggregate", this.Source); err != nil {
		return err
	}
	if !this.Function.IsValid() {
		return errors.Errorf("aggregate function %q is not supported", string(this.Function))
	}
	if err := this.validateOperand(); err != nil {
		return err
	}
	if this.Expr != nil {
		if nested := findRootOnlyNode(this.Expr); nested != "" {
			return errors.Errorf("%s cannot appear inside an aggregate expression", nested)
		}
	}
	return nil
}

func (this AggregateExpr) validateOperand() error {
	switch this.Function {
	case AggCount:
		if this.Field != "" || this.Expr != nil {
			return errors.New(`aggregate "count" counts rows and takes no operand`)
		}
	case AggCountDistinct:
		if this.Field == "" || this.Expr != nil {
			return errors.New(`aggregate "count_distinct" requires a source field name, not an expression`)
		}
	default:
		if (this.Field != "") == (this.Expr != nil) {
			return errors.Errorf(
				"aggregate %q requires exactly one operand: a source field or an expression",
				string(this.Function))
		}
	}
	return nil
}

func (this ExistsExpr) validate() error {
	if err := requireSourceEdge("exists", this.Source); err != nil {
		return err
	}
	if this.Filter == nil {
		return errors.New("exists requires a filter over the source schema")
	}
	return nil
}

func (this LookupExpr) validate() error {
	if err := requireSourceEdge("lookup", this.Source); err != nil {
		return err
	}
	if strings.TrimSpace(this.Field) == "" {
		return errors.New("lookup requires the source field to copy")
	}
	if len(this.OrderBy) == 0 {
		return errors.New("lookup requires at least one order_by entry")
	}
	for _, order := range this.OrderBy {
		if strings.TrimSpace(order.Field) == "" {
			return errors.New("lookup order_by entries require a field")
		}
	}
	return nil
}

func requireSourceEdge(kind string, source string) error {
	if strings.TrimSpace(source) == "" {
		return errors.Errorf("%s requires a source edge name", kind)
	}
	return nil
}
