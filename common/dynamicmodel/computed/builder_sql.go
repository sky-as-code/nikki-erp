package computed

import (
	"strings"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Chained constructors for the SQL-compiled kinds. Like the Go-evaluated builders they are pure —
// no registry lookups, no side effects — so a definition declares at package level and validates
// at schema finalize time:
//
//	Computed(false, computed.Aggregate("variants", computed.AggCount))
//	Computed(false, computed.Aggregate("lines", computed.AggSum,
//		computed.AggExpr(computed.Mul(computed.F("quantity"), computed.F("unit_price")))))
//	Computed(false, computed.Exists("invoices",
//		dmodel.NewSearchNode().NewCondition("status", dmodel.Equals, "posted")))
//	Computed(false, computed.Lookup("purchase_lines", "unit_price", computed.Desc("created_at")))

// AggregateOpt configures an optional part of an Aggregate definition.
type AggregateOpt func(*AggregateExpr)

// Aggregate aggregates a collection edge with one of the fixed functions. sum/avg/min/max need
// exactly one operand — AggField or AggExpr; count takes none; count_distinct takes AggField.
func Aggregate(source string, function AggregateFunction, opts ...AggregateOpt) Expr {
	node := AggregateExpr{Source: source, Function: function}
	for _, opt := range opts {
		opt(&node)
	}
	return node
}

// AggField aggregates a single column of the source schema.
func AggField(name string) AggregateOpt {
	return func(node *AggregateExpr) { node.Field = name }
}

// AggExpr aggregates an inner expression over source-schema fields, e.g. quantity * unit_price.
// Only the SQL-compilable subset is allowed: field refs, scalar literals, arithmetic operators,
// coalesce and nullif — finalize-time validation rejects anything else.
func AggExpr(expr Expr) AggregateOpt {
	return func(node *AggregateExpr) { node.Expr = expr }
}

// AggFilter narrows the aggregated rows with a search-node filter over the source schema.
func AggFilter(filter *dmodel.SearchNode) AggregateOpt {
	return func(node *AggregateExpr) { node.Filter = filter }
}

// AggContext whitelists request-context keys the filter references via Ctx.
func AggContext(keys ...string) AggregateOpt {
	return func(node *AggregateExpr) { node.Context = keys }
}

// AggDefault replaces a NULL aggregate result (sum/avg/min/max over zero rows).
func AggDefault(value any) AggregateOpt {
	return func(node *AggregateExpr) { node.Default = value }
}

// Exists is a boolean computed from the existence of source records matching the filter.
// Optional context keys whitelist Ctx references inside the filter.
func Exists(source string, filter *dmodel.SearchNode, contextKeys ...string) Expr {
	return ExistsExpr{Source: source, Filter: filter, Context: contextKeys}
}

// LookupOpt configures one part of a Lookup definition: ordering, filter, context or default.
type LookupOpt func(*LookupExpr)

// Lookup copies one scalar from the first source record after filter + ordering (LIMIT 1).
// At least one Asc/Desc ordering is required — validation enforces it.
func Lookup(source string, field string, opts ...LookupOpt) Expr {
	node := LookupExpr{Source: source, Field: field}
	for _, opt := range opts {
		opt(&node)
	}
	return node
}

// Asc appends an ascending ordering by a source-schema field.
func Asc(field string) LookupOpt {
	return func(node *LookupExpr) { node.OrderBy = append(node.OrderBy, OrderBy{Field: field}) }
}

// Desc appends a descending ordering by a source-schema field.
func Desc(field string) LookupOpt {
	return func(node *LookupExpr) { node.OrderBy = append(node.OrderBy, OrderBy{Field: field, Desc: true}) }
}

// LookupFilter narrows the candidate source records.
func LookupFilter(filter *dmodel.SearchNode) LookupOpt {
	return func(node *LookupExpr) { node.Filter = filter }
}

// LookupContext whitelists request-context keys the filter references via Ctx.
func LookupContext(keys ...string) LookupOpt {
	return func(node *LookupExpr) { node.Context = keys }
}

// LookupDefault replaces the NULL produced when no source record matches.
func LookupDefault(value any) LookupOpt {
	return func(node *LookupExpr) { node.Default = value }
}

// Ctx renders the placeholder a filter condition uses to bind a whitelisted request-context key,
// following the codebase's "${...}" substitution idiom:
//
//	dmodel.NewSearchNode().NewCondition("warehouse_id", dmodel.Equals, computed.Ctx("warehouse_id"))
//
// The whole-string form is required — a context value never concatenates into a larger string.
func Ctx(key string) string {
	return ctxPrefix + key + ctxSuffix
}

const (
	ctxPrefix = "${ctx."
	ctxSuffix = "}"
)

// CtxKeyOf recognizes a whole-string "${ctx.key}" filter value and extracts the key. Anything
// else — including a context placeholder embedded inside a longer string — is not a reference.
func CtxKeyOf(value any) (string, bool) {
	str, ok := value.(string)
	if !ok || len(str) <= len(ctxPrefix)+len(ctxSuffix) {
		return "", false
	}
	if !strings.HasPrefix(str, ctxPrefix) || !strings.HasSuffix(str, ctxSuffix) {
		return "", false
	}
	return str[len(ctxPrefix) : len(str)-len(ctxSuffix)], true
}

// looksLikePlaceholder flags any other "${...}" string so an unresolvable substitution fails
// validation instead of silently becoming a literal.
func looksLikePlaceholder(str string) bool {
	return strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}")
}
