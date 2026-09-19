package orm

import (
	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// This file is the seam between PgQueryBuilder and the enhanced builder in orm/advanced. The
// advanced builder owns its own projection and join planning but must produce predicates,
// ORDER BY expressions and literals byte-for-byte compatible with this package, so those
// primitives are exposed here rather than duplicated. Nothing in this file changes behaviour
// for callers of PgQueryBuilder itself.

// GraphRefResolver lets a caller intercept how a filter or order field path becomes a SQL
// reference. It is consulted before the join planner, so a path the planner would reject (a
// computed field, a deeper edge chain) can be answered by the caller. Returning ok=false hands
// the path back to the default planner resolution.
type GraphRefResolver interface {
	ResolveFilterRef(fieldName string) (field *dmodel.ModelField, sqlRef string, ok bool, err error)
	ResolveOrderRef(fieldName string) (field *dmodel.ModelField, sqlRef string, ok bool, err error)
}

// JoinSpec is one LEFT JOIN planned for a graph query.
type JoinSpec struct {
	TableWithAlias string
	OnExpr         string
}

// JoinPlan is the exported facade over the graph join planner: alias allocation, edge joins
// (including many-to-many junction hops), fan-out detection and the tenant predicates a
// junction join needs.
type JoinPlan struct {
	inner *joinPlanner
}

func (this *PgQueryBuilder) NewJoinPlan(registry *dmodel.SchemaRegistry, root *dmodel.ModelSchema) *JoinPlan {
	return &JoinPlan{inner: newJoinPlanner(this, registry, root)}
}

func (this *JoinPlan) Root() *dmodel.ModelSchema        { return this.inner.root }
func (this *JoinPlan) Registry() *dmodel.SchemaRegistry { return this.inner.registry }
func (this *JoinPlan) EnsureRootAliased()               { this.inner.ensureRootAliased() }
func (this *JoinPlan) RootAlias() string                { return this.inner.rootAlias }
func (this *JoinPlan) UsesJoins() bool                  { return this.inner.usesJoins() }
func (this *JoinPlan) NeedsDistinct() bool              { return this.inner.needsDistinct() }
func (this *JoinPlan) M2MTenantWheres() []string        { return this.inner.m2mTenantWheres }

func (this *JoinPlan) ResolveFieldSqlRef(field string, maxDots int) (*dmodel.ModelField, string, error) {
	return this.inner.resolveFieldSqlRef(field, maxDots)
}

// EnsureJoinedEdges joins every edge of the chain (cached per prefix) and returns the alias and
// schema of the last hop. An empty chain returns the root.
func (this *JoinPlan) EnsureJoinedEdges(edgeChain []string) (string, *dmodel.ModelSchema, error) {
	return this.inner.ensureJoinedEdges(edgeChain)
}

func (this *JoinPlan) Joins() []JoinSpec {
	out := make([]JoinSpec, len(this.inner.joins))
	for i, j := range this.inner.joins {
		out[i] = JoinSpec{TableWithAlias: j.tableWithAlias, OnExpr: j.onExpr}
	}
	return out
}

// RelationFansOut reports whether traversing rel can yield several destination rows per source row.
func RelationFansOut(rel dmodel.ModelRelation) bool { return relationFansOut(rel) }

// RelationByEdge finds the relation named edge on schema; the error is an unknown-field client error.
func RelationByEdge(schema *dmodel.ModelSchema, edge string) (dmodel.ModelRelation, error) {
	return relationByEdge(schema, edge)
}

// ApplyFromWithJoins writes FROM root [AS alias] plus every planned LEFT JOIN.
func (this *PgQueryBuilder) ApplyFromWithJoins(sb *sqlbuilder.SelectBuilder, plan *JoinPlan) {
	this.applyFromWithJoins(sb, plan.inner.root, plan.inner)
}

// CompileGraphPredicate renders the WHERE predicate of graph. resolver may be nil.
func (this *PgQueryBuilder) CompileGraphPredicate(
	plan *JoinPlan, resolver GraphRefResolver, language *model.LanguageCode,
	sb *sqlbuilder.SelectBuilder, graph *dmodel.SearchGraph,
) (string, ft.ClientErrors, error) {
	if graph == nil {
		return "", nil, nil
	}
	ctx := &graphSelectCtx{planner: plan.inner, language: language, resolver: resolver}
	return this.graphExpression(ctx, plan.inner.root, sb, graph.GetCondition(), graph.GetAnd(), graph.GetOr())
}

// CompileOrderExprs renders "ref ASC|DESC" items for order. resolver may be nil.
func (this *PgQueryBuilder) CompileOrderExprs(
	plan *JoinPlan, resolver GraphRefResolver, language *model.LanguageCode, order dmodel.SearchOrder,
) ([]string, error) {
	ctx := &graphSelectCtx{planner: plan.inner, language: language, resolver: resolver}
	return this.orderExprs(ctx, plan.inner.root, order)
}

// ConvertValue converts a filter value to the Go type the field's column expects.
func (this *PgQueryBuilder) ConvertValue(field *dmodel.ModelField, value any) (any, ft.ClientErrors, error) {
	return this.convertValue(field, value)
}

func (this *PgQueryBuilder) TableExpression(schema *dmodel.ModelSchema) string {
	return this.tableExpression(schema)
}

func (this *PgQueryBuilder) ApplyPagination(sb *sqlbuilder.SelectBuilder, page, size int) {
	this.applyPagination(sb, page, size)
}

// BuildAggregateSelectExpr renders a COUNT|MAX|MIN|AVG|SUM(field) select token.
func (this *PgQueryBuilder) BuildAggregateSelectExpr(
	plan *JoinPlan, token SelectColumn, funcUpper, inner string,
) (string, ft.ClientErrors, error) {
	return buildAggregateSelectExpr(plan.inner, token, funcUpper, inner)
}

func PgQuote(s string) string                            { return pgQuote(s) }
func PgStringLiteral(value string) string                { return pgStringLiteral(value) }
func SqlLiteral(v any) (string, error)                   { return sqlLiteralForLinkedSubquery(v) }
func Interpolate(sql string, args []any) (string, error) { return interpolate(sql, args) }
func IsLangJsonField(field *dmodel.ModelField) bool      { return isLangJsonField(field) }

func LangJsonLocalizedExpr(field *dmodel.ModelField, sqlRef string, language *model.LanguageCode) string {
	return langJsonLocalizedExpr(field, sqlRef, language)
}

func ParseAllowedAggregate(col SelectColumn) (funcUpper string, inner string, ok bool) {
	return parseAllowedAggregate(col)
}
func SelectColumnIsDistinct(col SelectColumn) bool     { return selectColumnHasDistinctPrefix(col) }
func SelectColumnPath(col SelectColumn) string         { return selectColumnStripDistinctPrefix(col) }
func AnySelectColumnDistinct(cols []SelectColumn) bool { return anySelectColumnDistinct(cols) }
func (this SelectColumn) JoinPlanningPath() string     { return this.joinPlanningPath() }
func (this SelectColumn) Raw() string                  { return this.rawString() }
func OrderRefsForDistinct(isDistinct bool, orderExprs []string) []string {
	return orderRefsForDistinct(isDistinct, orderExprs)
}

// SqlGraphOutcome shapes a builder result into the QueryBuilder return contract, mapping
// wrapped client errors out of err.
func SqlGraphOutcome(sql string, cErrs ft.ClientErrors, err error) (*string, *ft.ClientErrors, error) {
	return stringSqlGraphOutcome(sql, cErrs, err)
}

// WrapClientErrors carries client errors through an error return; SqlGraphOutcome unwraps them.
func WrapClientErrors(c ft.ClientErrors) error { return wrapClientSqlErrors(c) }

// ErrUnknownField is the error a resolver returns for a path that names no field; SqlGraphOutcome
// maps it to err_unknown_schema_field.
func ErrUnknownField(field string) error {
	return errors.Wrap(&errClientUnknownField{Field: field}, "ErrUnknownField")
}

func ClientErrorsGraphFieldTooDeep(field string, maxDots int) ft.ClientErrors {
	return clientErrorsGraphFieldTooDeep(field, maxDots)
}
func ClientErrorsInvalidGraphFieldPath(field string) ft.ClientErrors {
	return clientErrorsInvalidGraphFieldPath(field)
}
func ClientErrorsUnknownField(field string) ft.ClientErrors { return clientErrorsUnknownField(field) }
func ClientErrorsVirtualFieldUnavailable(field string) ft.ClientErrors {
	return clientErrorsVirtualFieldUnavailable(field)
}
func ClientErrorsFieldNotSortable(field string) ft.ClientErrors {
	return clientErrorsFieldNotSortable(field)
}
func ClientErrorsTooManySqlComputedFields(limit int) ft.ClientErrors {
	return clientErrorsTooManySqlComputedFields(limit)
}
func ClientErrorsInvalidSelectAggregate(token string) ft.ClientErrors {
	return clientErrorsInvalidSelectAggregate(token)
}
func ClientErrorsRegistryRequiredForGraph() ft.ClientErrors {
	return clientErrorsRegistryRequiredForGraph()
}
func ClientErrorsUnsupportedRelationInGraph(edge string, relType dmodel.RelationType) ft.ClientErrors {
	return clientErrorsUnsupportedRelationInGraph(edge, relType)
}
