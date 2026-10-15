package computed

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// sqlRenderableFunctions are the builtins the advanced query builder renders to SQL. It is the
// single list both the builder and the schema metadata consult, so a field is never advertised
// as filterable while the builder refuses it. Date/time builtins are absent on purpose: their
// Go semantics (calendar arithmetic on the model wrappers) have no rendering yet.
var sqlRenderableFunctions = map[string]bool{
	"coalesce": true,
	"nullif":   true,
	"concat":   true,
	"lower":    true,
	"upper":    true,
	"trim":     true,
	"length":   true,
	"abs":      true,
	"ceil":     true,
	"floor":    true,
	"round":    true,
}

// IsSqlRenderableFunction reports whether a builtin has an SQL rendering.
func IsSqlRenderableFunction(name string) bool {
	return sqlRenderableFunctions[name]
}

// IsSqlRenderable reports whether a computed field can be rendered in SQL on its own schema, and
// therefore used in a filter or an order. A function-kind field is evaluated in Go and never is;
// an expression is only when every builtin it calls and every computed field it reads is.
//
// The answer describes the field on its own schema. A path reaching it through a to-many edge,
// or exceeding the path depth caps, is still refused by the query builder at request time.
func IsSqlRenderable(schemaName string, fieldName string) bool {
	schemaPlan := PlanFor(schemaName)
	if schemaPlan == nil {
		return false
	}
	return isFieldSqlRenderable(schemaPlan, fieldName, 0)
}

func isFieldSqlRenderable(schemaPlan *SchemaPlan, fieldName string, exprDepth int) bool {
	fieldPlan := schemaPlan.Fields[fieldName]
	if fieldPlan == nil || fieldPlan.Def == nil {
		return false
	}
	switch fieldPlan.Def.Kind {
	case ComputeRelated, ComputeAggregate, ComputeExists, ComputeLookup:
		return true
	case ComputeExpression:
		// Mirrors the builder, which refuses an expression nested this deep in other expressions.
		if exprDepth >= ActiveLimits().MaxComputedDependencyDepth {
			return false
		}
		if !callsOnlySqlRenderableFunctions(fieldPlan.Def.Expression) {
			return false
		}
		for _, dependency := range fieldPlan.ComputedDeps {
			if !isFieldSqlRenderable(schemaPlan, dependency, exprDepth+1) {
				return false
			}
		}
		return true
	}
	return false
}

func callsOnlySqlRenderableFunctions(expr Expr) bool {
	switch node := expr.(type) {
	case BinaryExpr:
		return callsOnlySqlRenderableFunctions(node.Left) && callsOnlySqlRenderableFunctions(node.Right)
	case UnaryExpr:
		return callsOnlySqlRenderableFunctions(node.Operand)
	case FunctionExpr:
		if !IsSqlRenderableFunction(node.Name) {
			return false
		}
		for _, arg := range node.Args {
			if !callsOnlySqlRenderableFunctions(arg) {
				return false
			}
		}
		return true
	case CaseExpr:
		for _, branch := range node.Whens {
			if !callsOnlySqlRenderableFunctions(branch.When) || !callsOnlySqlRenderableFunctions(branch.Then) {
				return false
			}
		}
		return node.Else == nil || callsOnlySqlRenderableFunctions(node.Else)
	}
	return true
}

func init() {
	dmodel.RegisterComputedSqlRenderable(IsSqlRenderable)
}
