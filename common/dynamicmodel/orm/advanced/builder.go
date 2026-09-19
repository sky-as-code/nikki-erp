// Package advanced holds AdvancedPgQueryBuilder, the enhanced PostgreSQL graph query builder.
//
// It keeps every capability of orm.PgQueryBuilder (DDL, DML, unique and exists helpers and
// predefined predicates come straight from the embedded builder) and replaces the three graph
// queries — select, count, exists — with a planner that:
//
//   - materialises each SQL-computed field (aggregate / exists / lookup) once, as a LEFT JOIN
//     LATERAL the projection, WHERE and ORDER BY all reference, so a field used in both the
//     filter and the selection is computed a single time;
//   - accepts computed fields of every kind except Go functions in filters and sort orders,
//     compiling expression fields to SQL and delegating related fields to their joined leaf;
//   - projects to-one edge fields through joins and to-many edges through jsonb aggregation
//     inside the root statement, so a page never costs one follow-up query per row;
//   - resolves computed fields reached through a to-one edge against the edge's own alias.
//
// Predicates, ORDER BY items and literals are produced by the exported seam in
// orm/advanced_api.go, so the two builders never disagree on escaping or operator semantics.
package advanced

import (
	"fmt"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

var _ orm.QueryBuilder = (*AdvancedPgQueryBuilder)(nil)

type AdvancedPgQueryBuilder struct {
	*orm.PgQueryBuilder
}

func NewAdvancedPgQueryBuilder() orm.QueryBuilder {
	return &AdvancedPgQueryBuilder{PgQueryBuilder: orm.NewPgQueryBuilder().(*orm.PgQueryBuilder)}
}

func (this *AdvancedPgQueryBuilder) SqlSelectGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildSelect(schema, registry, graph, opts)
	return orm.SqlGraphOutcome(sql, cErrs, err)
}

func (this *AdvancedPgQueryBuilder) SqlCountGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildCount(schema, registry, graph, opts)
	return orm.SqlGraphOutcome(sql, cErrs, err)
}

func (this *AdvancedPgQueryBuilder) SqlExistsGraph(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph,
) (*string, *ft.ClientErrors, error) {
	sql, cErrs, err := this.buildExists(schema, registry, graph)
	return orm.SqlGraphOutcome(sql, cErrs, err)
}

func (this *AdvancedPgQueryBuilder) buildSelect(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (string, ft.ClientErrors, error) {
	plan, err := this.newPlan(schema, registry, graph, opts, planSelect)
	if err != nil {
		return "", nil, err
	}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	isDistinct := orm.AnySelectColumnDistinct(opts.Columns) || plan.join.NeedsDistinct()
	if isDistinct {
		sb.Distinct()
	}
	// Under SELECT DISTINCT every ORDER BY expression must be projected, so the order is
	// resolved before the projection is written.
	var orderExprs []string
	if graph != nil {
		orderExprs, err = this.CompileOrderExprs(plan.join, plan.forUse(useOrder), opts.Language, graph.GetOrder())
		if err != nil {
			return "", nil, err
		}
	}
	if err := plan.applySelectColumns(sb, opts.Columns, orm.OrderRefsForDistinct(isDistinct, orderExprs)...); err != nil {
		return "", nil, err
	}
	predicate, cErrs, err := this.CompileGraphPredicate(plan.join, plan.forUse(useFilter), opts.Language, sb, graph)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	plan.applyFrom(sb, lateralUseAll)
	if predicate != "" {
		sb.Where(predicate)
	}
	if len(orderExprs) > 0 {
		sb.OrderBy(orderExprs...)
	}
	this.ApplyPagination(sb, opts.Page, opts.Size)
	return build(sb, "buildSelect")
}

func (this *AdvancedPgQueryBuilder) buildExists(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph,
) (string, ft.ClientErrors, error) {
	plan, err := this.newPlan(schema, registry, graph, orm.SqlSelectGraphOpts{}, planExists)
	if err != nil {
		return "", nil, err
	}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("1")
	predicate, cErrs, err := this.CompileGraphPredicate(plan.join, plan.forUse(useFilter), nil, sb, graph)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	plan.applyFrom(sb, useFilter)
	if predicate != "" {
		sb.Where(predicate)
	}
	inner, _, err := build(sb, "buildExists")
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("SELECT EXISTS (%s)", inner), nil, nil
}

func (this *AdvancedPgQueryBuilder) buildCount(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (string, ft.ClientErrors, error) {
	plan, err := this.newPlan(schema, registry, graph, opts, planCount)
	if err != nil {
		return "", nil, err
	}
	if orm.AnySelectColumnDistinct(opts.Columns) || plan.join.NeedsDistinct() {
		return this.buildCountDistinct(plan, graph, opts)
	}
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select("COUNT(*)")
	predicate, cErrs, err := this.CompileGraphPredicate(plan.join, plan.forUse(useFilter), opts.Language, sb, graph)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	plan.applyFrom(sb, useFilter)
	if predicate != "" {
		sb.Where(predicate)
	}
	return build(sb, "buildCount")
}

// buildCountDistinct counts over the same DISTINCT projection the list query uses, so Total
// matches the rows beside it when a fan-out join is in play.
func (this *AdvancedPgQueryBuilder) buildCountDistinct(
	plan *queryPlan, graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (string, ft.ClientErrors, error) {
	inner := sqlbuilder.PostgreSQL.NewSelectBuilder()
	inner.Distinct()
	if err := plan.applySelectColumns(inner, opts.Columns); err != nil {
		return "", nil, err
	}
	predicate, cErrs, err := this.CompileGraphPredicate(plan.join, plan.forUse(useFilter), opts.Language, inner, graph)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	plan.applyFrom(inner, useFilter|useSelect)
	if predicate != "" {
		inner.Where(predicate)
	}
	innerSql, _, err := build(inner, "buildCountDistinct")
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS _distinct_count", innerSql), nil, nil
}

func build(sb *sqlbuilder.SelectBuilder, op string) (string, ft.ClientErrors, error) {
	raw, args := sb.Build()
	out, err := orm.Interpolate(raw, args)
	if err != nil {
		return "", nil, errors.Wrap(err, op+": interpolate")
	}
	return out, nil, nil
}
