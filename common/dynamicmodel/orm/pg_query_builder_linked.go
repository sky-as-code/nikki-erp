package orm

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func (this *PgQueryBuilder) linkedNotLinkedEdgePredicate(
	ctx *graphSelectCtx, schema *dmodel.ModelSchema, fieldName string, op dmodel.Operator, value any,
) (string, ft.ClientErrors, error) {
	if ctx == nil || ctx.planner == nil {
		return "", nil, errors.New("linkedNotLinkedEdgePredicate: planner required")
	}
	if ctx.planner.registry == nil {
		return "", nil, errors.New("linkedNotLinkedEdgePredicate: schema registry required")
	}
	edge := strings.TrimSpace(fieldName)
	if edge == "" || strings.Contains(edge, ".") {
		return "", clientErrorsLinkedEdgeField(), nil
	}
	rel, err := relationByEdge(schema, edge)
	if err != nil {
		return "", nil, err
	}
	switch rel.RelationType {
	case dmodel.RelationTypeOneToMany:
		return this.linkedO2MSubqueryPredicate(ctx, schema, rel, op, value)
	case dmodel.RelationTypeManyToMany:
		return this.linkedM2MSubqueryPredicate(ctx, schema, rel, op, value)
	default:
		return "", clientErrorsLinkedManyEdgeOnly(edge), nil
	}
}

func clientErrorsLinkedEdgeField() ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError("graph.condition", ft.ErrorKey("err_graph_linked_operator_field"),
			"linked/not_linked require a single edge name (no dot path)"),
	}
}

func clientErrorsLinkedManyEdgeOnly(edge string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(edge, ft.ErrorKey("err_graph_linked_operator_many_only"),
			"linked/not_linked are only valid on one:many or many:many edges"),
	}
}

func clientErrorsLinkedCompositePk() ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewAnonymousValidationError(ft.ErrorKey("err_graph_linked_composite_pk"),
			"linked/not_linked: composite primary key is not supported yet", nil),
	}
}

func rootPkSqlExpr(p *joinPlanner, pkCol string) string {
	q := pgQuote(pkCol)
	if p.rootAlias != "" {
		return fmt.Sprintf("%s.%s", p.rootAlias, q)
	}
	return q
}

func sqlLiteralForLinkedSubquery(v any) (string, error) {
	switch x := v.(type) {
	case string:
		return pgStringLiteral(x), nil
	case bool:
		if x {
			return "TRUE", nil
		}
		return "FALSE", nil
	case int:
		return fmt.Sprintf("%d", x), nil
	case int32:
		return fmt.Sprintf("%d", x), nil
	case int64:
		return fmt.Sprintf("%d", x), nil
	case uint:
		return fmt.Sprintf("%d", x), nil
	case uint32:
		return fmt.Sprintf("%d", x), nil
	case uint64:
		return fmt.Sprintf("%d", x), nil
	case float32:
		return fmt.Sprintf("%g", x), nil
	case float64:
		return fmt.Sprintf("%g", x), nil
	case decimal.Decimal:
		return x.String(), nil
	case *decimal.Decimal:
		if x == nil {
			return "", errors.New("linked: null decimal value")
		}
		return x.String(), nil
	default:
		return pgStringLiteral(fmt.Sprint(v)), nil
	}
}

func (this *PgQueryBuilder) linkedO2MSubqueryPredicate(
	ctx *graphSelectCtx, root *dmodel.ModelSchema, rel dmodel.ModelRelation, op dmodel.Operator, value any,
) (string, ft.ClientErrors, error) {
	p := ctx.planner
	pairs := rel.EffectiveForeignKeys()
	rootPks := root.PrimaryKeys()
	destSch := p.registry.Get(rel.DestSchemaName)
	if destSch == nil {
		return "", nil, errors.Errorf("linked O2M: destination schema %q not in registry", rel.DestSchemaName)
	}
	if len(rootPks) != 1 || len(pairs) != 1 {
		return "", clientErrorsLinkedCompositePk(), nil
	}
	destPks := destSch.PrimaryKeys()
	if len(destPks) != 1 {
		return "", clientErrorsLinkedCompositePk(), nil
	}
	childPk := destPks[0]
	fkCol := pairs[0].FkColumn
	pkField, ok := destSch.Column(childPk)
	if !ok {
		return "", nil, errors.Errorf("linked O2M: unknown child pk %q", childPk)
	}
	converted, cErrs, err := this.convertValue(pkField, value)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	lit, err := sqlLiteralForLinkedSubquery(converted)
	if err != nil {
		return "", nil, err
	}
	wheres := fmt.Sprintf("%s = %s", pgQuote(childPk), lit)
	if tk := root.TenantKey(); tk != "" {
		if dtk := destSch.TenantKey(); dtk != "" {
			wheres = fmt.Sprintf("%s AND %s = %s", wheres, pgQuote(dtk), rootPkSqlExpr(p, tk))
		}
	}
	sub := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		pgQuote(fkCol), this.tableExpression(destSch), wheres,
	)
	return formatLinkedInOrNotIn(rootPkSqlExpr(p, rootPks[0]), sub, op), nil, nil
}

func (this *PgQueryBuilder) linkedM2MSubqueryPredicate(
	ctx *graphSelectCtx, root *dmodel.ModelSchema, rel dmodel.ModelRelation, op dmodel.Operator, value any,
) (string, ft.ClientErrors, error) {
	p := ctx.planner
	if rel.M2mThroughSchemaName == "" || rel.M2mSrcFieldPrefix == "" || rel.M2mDestFieldPrefix == "" {
		return "", nil, errors.New("linked M2M: junction metadata incomplete")
	}
	throughSch := p.registry.Get(rel.M2mThroughSchemaName)
	destSch := p.registry.Get(rel.DestSchemaName)
	if throughSch == nil || destSch == nil {
		return "", nil, errors.New("linked M2M: through or peer schema not in registry")
	}
	rootPks := root.PrimaryKeys()
	peerPks := destSch.PrimaryKeys()
	if len(rootPks) != 1 || len(peerPks) != 1 {
		return "", clientErrorsLinkedCompositePk(), nil
	}
	srcCol := dmodel.PrefixedThroughColumn(rel.M2mSrcFieldPrefix, rootPks[0])
	peerCol := dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, peerPks[0])
	pkField, ok := destSch.Column(peerPks[0])
	if !ok {
		return "", nil, errors.Errorf("linked M2M: unknown peer pk %q", peerPks[0])
	}
	converted, cErrs, err := this.convertValue(pkField, value)
	if err != nil {
		return "", nil, err
	}
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	lit, err := sqlLiteralForLinkedSubquery(converted)
	if err != nil {
		return "", nil, err
	}
	const alias = "th"
	wheres := fmt.Sprintf("%s.%s = %s", alias, pgQuote(peerCol), lit)
	// Match appendManyToManyJoin: junction tenant column name equals parent TenantKey(), not src_prefix+tenant.
	if tk := root.TenantKey(); tk != "" {
		wheres = fmt.Sprintf("%s AND %s.%s = %s", wheres, alias, pgQuote(tk), rootPkSqlExpr(p, tk))
	}
	sub := fmt.Sprintf(
		"SELECT %s.%s FROM %s AS %s WHERE %s",
		alias, pgQuote(srcCol), this.tableExpression(throughSch), alias, wheres,
	)
	return formatLinkedInOrNotIn(rootPkSqlExpr(p, rootPks[0]), sub, op), nil, nil
}

func formatLinkedInOrNotIn(rootExpr, subquery string, op dmodel.Operator) string {
	if op == dmodel.NotLinked {
		return fmt.Sprintf("%s NOT IN (%s)", rootExpr, subquery)
	}
	return fmt.Sprintf("%s IN (%s)", rootExpr, subquery)
}
