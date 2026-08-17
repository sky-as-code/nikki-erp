package orm

import (
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// Correlated-subquery emission for the SQL-computed kinds (aggregate / exists / lookup). One
// computed field becomes exactly one scalar subquery in the SELECT projection — never a JOIN +
// GROUP BY on the root query, which would fan out and mis-aggregate when two collection edges
// are requested together (spec §13).
//
// Inside the subquery every column name is unqualified and therefore resolves to the source
// table (the innermost FROM); the only qualified references are to the root row through its
// alias, which is why the caller must have aliased the root. Filter predicates and context
// values flow through the same convertValue + interpolation path as every other query, so a
// hostile value survives only as an escaped literal.

// archiveFieldName mirrors baserepo's is_archived injection on root queries: when the source
// schema declares the column, archived source rows never count toward a computed value.
const archiveFieldName = "is_archived"

// computedSelectExpr resolves a projected virtual-scalar path against the schema's computed
// plans. When the field is SQL-computed it returns the aliased subquery expression and forces
// the root alias the correlation needs; a Go-computed or plain virtual scalar returns
// emitted=false so the projection keeps skipping it.
func (this *PgQueryBuilder) computedSelectExpr(
	planner *joinPlanner, path string, ctxValues map[string]any, count *int,
) (string, bool, error) {
	schemaPlan := computed.PlanFor(planner.root.Name())
	if schemaPlan == nil {
		return "", false, nil
	}
	fieldPlan := schemaPlan.Fields[path]
	if fieldPlan == nil || fieldPlan.SqlSource == nil {
		return "", false, nil
	}
	if planner.registry == nil {
		return "", false, errors.Errorf(
			"computedSelectExpr: field %q needs the schema registry to compile its subquery", path)
	}
	*count++
	if limit := computed.ActiveLimits().MaxSqlComputedFieldsPerRequest; *count > limit {
		return "", false, wrapClientSqlErrors(clientErrorsTooManySqlComputedFields(limit))
	}
	planner.ensureRootAliased()
	subquery, cErrs, err := this.ComputedSubqueryExpr(
		planner.registry, planner.root, planner.rootAlias, fieldPlan, ctxValues)
	if err != nil {
		return "", false, err
	}
	if len(cErrs) > 0 {
		return "", false, wrapClientSqlErrors(cErrs)
	}
	return subquery + " AS " + pgQuote(path), true, nil
}

func clientErrorsTooManySqlComputedFields(limit int) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError("fields", ft.ErrorKey("err_too_many_sql_computed_fields"),
			fmt.Sprintf("request projects more than %d SQL-computed fields", limit)),
	}
}

// ComputedSubqueryExpr renders the correlated subquery for one SQL-computed field, including
// the COALESCE(default) wrapper when the definition declares one.
func (this *PgQueryBuilder) ComputedSubqueryExpr(
	registry *dmodel.SchemaRegistry, root *dmodel.ModelSchema, rootAlias string,
	plan *computed.FieldPlan, ctxValues map[string]any,
) (string, ft.ClientErrors, error) {
	if rootAlias == "" {
		return "", nil, errors.New("ComputedSubqueryExpr: the root query must be aliased for correlation")
	}
	if plan == nil || plan.SqlSource == nil {
		return "", nil, errors.New("ComputedSubqueryExpr: plan is not an SQL-computed field plan")
	}
	sourceSchema := registry.Get(plan.SqlSource.SourceSchemaName)
	if sourceSchema == nil {
		return "", nil, errors.Errorf("ComputedSubqueryExpr: source schema %q not in registry",
			plan.SqlSource.SourceSchemaName)
	}
	emit := &computedSubqueryEmit{
		builder: this, registry: registry, root: root, rootAlias: rootAlias,
		plan: plan, sourceSchema: sourceSchema, ctxValues: ctxValues,
	}
	return emit.render()
}

type computedSubqueryEmit struct {
	builder      *PgQueryBuilder
	registry     *dmodel.SchemaRegistry
	root         *dmodel.ModelSchema
	rootAlias    string
	plan         *computed.FieldPlan
	sourceSchema *dmodel.ModelSchema
	ctxValues    map[string]any
}

func (this *computedSubqueryEmit) render() (string, ft.ClientErrors, error) {
	switch this.plan.Def.Kind {
	case computed.ComputeAggregate:
		return this.renderAggregate(this.plan.Def.Aggregate)
	case computed.ComputeExists:
		return this.renderExists(this.plan.Def.Exists)
	case computed.ComputeLookup:
		return this.renderLookup(this.plan.Def.Lookup)
	}
	return "", nil, errors.Errorf("ComputedSubqueryExpr: kind %q is not SQL-computed", this.plan.Def.Kind)
}

func (this *computedSubqueryEmit) renderAggregate(node *computed.AggregateExpr) (string, ft.ClientErrors, error) {
	selectExpr, err := aggregateSelectExpr(node)
	if err != nil {
		return "", nil, err
	}
	sql, cErrs, err := this.subquerySql(selectExpr, node.Filter, nil)
	if err != nil || len(cErrs) > 0 {
		return "", cErrs, err
	}
	return this.wrapDefault(sql, node.Default)
}

func (this *computedSubqueryEmit) renderExists(node *computed.ExistsExpr) (string, ft.ClientErrors, error) {
	sql, cErrs, err := this.subquerySql("1", node.Filter, nil)
	if err != nil || len(cErrs) > 0 {
		return "", cErrs, err
	}
	return "EXISTS " + sql, nil, nil
}

func (this *computedSubqueryEmit) renderLookup(node *computed.LookupExpr) (string, ft.ClientErrors, error) {
	sql, cErrs, err := this.subquerySql(pgQuote(node.Field), node.Filter, node.OrderBy)
	if err != nil || len(cErrs) > 0 {
		return "", cErrs, err
	}
	return this.wrapDefault(sql, node.Default)
}

// subquerySql assembles "(SELECT <expr> FROM <source> WHERE <correlation AND scope AND filter>
// [ORDER BY ... LIMIT 1])" and interpolates it into a complete SQL fragment.
func (this *computedSubqueryEmit) subquerySql(
	selectExpr string, filter *dmodel.SearchNode, orderBy []computed.OrderBy,
) (string, ft.ClientErrors, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(selectExpr)
	sb.From(this.builder.tableExpression(this.sourceSchema))

	wheres, cErrs, err := this.whereClauses(sb, filter)
	if err != nil || len(cErrs) > 0 {
		return "", cErrs, err
	}
	sb.Where(wheres...)
	if len(orderBy) > 0 {
		sb.OrderBy(lookupOrderExprs(orderBy)...)
		sb.Limit(1)
	}

	sql, args := sb.Build()
	out, err := interpolate(sql, args)
	if err != nil {
		return "", nil, errors.Wrap(err, "computed subquery: interpolate")
	}
	return "(" + out + ")", nil, nil
}

// whereClauses emits, in order: the correlation to the root row, the tenant correlation, the
// archive scope, and the definition's filter.
func (this *computedSubqueryEmit) whereClauses(
	sb *sqlbuilder.SelectBuilder, filter *dmodel.SearchNode,
) ([]string, ft.ClientErrors, error) {
	wheres := []string{this.correlationClause()}
	if tenantClause := this.tenantClause(); tenantClause != "" {
		wheres = append(wheres, tenantClause)
	}
	if _, ok := this.sourceSchema.Column(archiveFieldName); ok {
		wheres = append(wheres, fmt.Sprintf("%s = FALSE", pgQuote(archiveFieldName)))
	}
	predicate, cErrs, err := this.filterPredicate(sb, filter)
	if err != nil || len(cErrs) > 0 {
		return nil, cErrs, err
	}
	if predicate != "" {
		wheres = append(wheres, predicate)
	}
	return wheres, nil, nil
}

// correlationClause pairs source rows with the current root row. one:many correlates the FK
// column directly; many:many goes through the junction table with the same shape
// linkedM2MSubqueryPredicate uses.
func (this *computedSubqueryEmit) correlationClause() string {
	source := this.plan.SqlSource
	if !source.Many {
		return fmt.Sprintf("%s = %s.%s",
			pgQuote(source.SourceFkColumn), this.rootAlias, pgQuote(source.RootRefColumn))
	}
	throughSchema := this.registry.Get(source.ThroughSchemaName)
	throughTable := source.ThroughSchemaName
	if throughSchema != nil {
		throughTable = strings.TrimSpace(this.builder.tableExpression(throughSchema))
	}
	membership := fmt.Sprintf("SELECT %s FROM %s WHERE %s = %s.%s",
		pgQuote(source.ThroughDestColumn), throughTable,
		pgQuote(source.ThroughSrcColumn), this.rootAlias, pgQuote(source.RootPkColumn))
	if tenantKey := this.root.TenantKey(); tenantKey != "" {
		// Junction tenant column name equals the parent TenantKey(); see appendManyToManyJoin.
		membership = fmt.Sprintf("%s AND %s = %s.%s",
			membership, pgQuote(tenantKey), this.rootAlias, pgQuote(tenantKey))
	}
	return fmt.Sprintf("%s IN (%s)", pgQuote(source.SourcePkColumn), membership)
}

func (this *computedSubqueryEmit) tenantClause() string {
	if this.plan.SqlSource.Many {
		return "" // The junction membership already carries the tenant correlation.
	}
	rootTenant := this.root.TenantKey()
	sourceTenant := this.sourceSchema.TenantKey()
	if rootTenant == "" || sourceTenant == "" {
		return ""
	}
	return fmt.Sprintf("%s = %s.%s", pgQuote(sourceTenant), this.rootAlias, pgQuote(rootTenant))
}

// filterPredicate compiles the definition's filter against the SOURCE schema through the same
// graphExpression machinery every search uses, after substituting "${ctx.key}" values.
func (this *computedSubqueryEmit) filterPredicate(
	sb *sqlbuilder.SelectBuilder, filter *dmodel.SearchNode,
) (string, ft.ClientErrors, error) {
	if filter == nil {
		return "", nil, nil
	}
	bound, cErrs := substituteContextValues(filter, this.ctxValues)
	if len(cErrs) > 0 {
		return "", cErrs, nil
	}
	return this.builder.graphExpression(
		nil, this.sourceSchema, sb, bound.GetCondition(), bound.GetAnd(), bound.GetOr())
}

func (this *computedSubqueryEmit) wrapDefault(sql string, defaultValue any) (string, ft.ClientErrors, error) {
	if defaultValue == nil {
		return sql, nil, nil
	}
	literal, err := sqlLiteralForLinkedSubquery(defaultValue)
	if err != nil {
		return "", nil, errors.Wrap(err, "computed default")
	}
	return fmt.Sprintf("COALESCE(%s, %s)", sql, literal), nil, nil
}

func lookupOrderExprs(orderBy []computed.OrderBy) []string {
	exprs := make([]string, len(orderBy))
	for i, order := range orderBy {
		direction := " ASC"
		if order.Desc {
			direction = " DESC"
		}
		exprs[i] = pgQuote(order.Field) + direction
	}
	return exprs
}

// substituteContextValues deep-copies the filter with every whole-string "${ctx.key}" value
// replaced by the bound request value. A referenced key with no bound value is a client error —
// an unfiltered fallback would silently widen the computation's scope.
func substituteContextValues(
	node *dmodel.SearchNode, ctxValues map[string]any,
) (*dmodel.SearchNode, ft.ClientErrors) {
	if condition := node.GetCondition(); condition != nil {
		bound, cErrs := substituteConditionValues(condition, ctxValues)
		if len(cErrs) > 0 {
			return nil, cErrs
		}
		return dmodel.NewSearchNode().Condition(bound), nil
	}
	if children := node.GetAnd(); len(children) > 0 {
		bound, cErrs := substituteChildNodes(children, ctxValues)
		return dmodel.NewSearchNode().And(bound...), cErrs
	}
	if children := node.GetOr(); len(children) > 0 {
		bound, cErrs := substituteChildNodes(children, ctxValues)
		return dmodel.NewSearchNode().Or(bound...), cErrs
	}
	return dmodel.NewSearchNode(), nil
}

func substituteChildNodes(
	children []dmodel.SearchNode, ctxValues map[string]any,
) ([]dmodel.SearchNode, ft.ClientErrors) {
	bound := make([]dmodel.SearchNode, len(children))
	for i := range children {
		child, cErrs := substituteContextValues(&children[i], ctxValues)
		if len(cErrs) > 0 {
			return nil, cErrs
		}
		bound[i] = *child
	}
	return bound, nil
}

func substituteConditionValues(
	condition dmodel.Condition, ctxValues map[string]any,
) (dmodel.Condition, ft.ClientErrors) {
	bound := make(dmodel.Condition, len(condition))
	copy(bound, condition)
	for i := 2; i < len(bound); i++ {
		key, ok := computed.CtxKeyOf(bound[i])
		if !ok {
			continue
		}
		value, exists := ctxValues[key]
		if !exists {
			return nil, ft.ClientErrors{
				*ft.NewValidationError(key, ft.ErrorKey("err_computed_context_missing"),
					fmt.Sprintf("computed field requires context value %q", key)),
			}
		}
		bound[i] = value
	}
	return bound, nil
}

// aggregateSelectExpr renders the aggregate function call: COUNT(*), COUNT(DISTINCT col), or
// FN(operand) where the operand is a column or the SQL-compiled inner expression.
func aggregateSelectExpr(node *computed.AggregateExpr) (string, error) {
	switch node.Function {
	case computed.AggCount:
		return "COUNT(*)", nil
	case computed.AggCountDistinct:
		return "COUNT(DISTINCT " + pgQuote(node.Field) + ")", nil
	}
	operand := pgQuote(node.Field)
	if node.Expr != nil {
		compiled, err := computedInnerSqlExpr(node.Expr)
		if err != nil {
			return "", err
		}
		operand = compiled
	}
	return strings.ToUpper(string(node.Function)) + "(" + operand + ")", nil
}

// computedInnerSqlExpr compiles the restricted inner-expression subset to SQL. Only the shapes
// finalize-time validation admits can appear (field refs, scalar literals, arithmetic, negate,
// coalesce/nullif); anything else is an internal error, not a client one.
func computedInnerSqlExpr(expr computed.Expr) (string, error) {
	switch node := expr.(type) {
	case computed.FieldExpr:
		return pgQuote(node.Name), nil
	case computed.LiteralExpr:
		if node.Value == nil {
			return "NULL", nil
		}
		return sqlLiteralForLinkedSubquery(node.Value)
	case computed.BinaryExpr:
		return computedInnerBinarySql(node)
	case computed.UnaryExpr:
		if node.Op != computed.OpNegate {
			return "", errors.Errorf("computed inner expression: operator %q cannot compile to SQL", node.Op)
		}
		operand, err := computedInnerSqlExpr(node.Operand)
		if err != nil {
			return "", err
		}
		return "(- " + operand + ")", nil
	case computed.FunctionExpr:
		return computedInnerFunctionSql(node)
	}
	return "", errors.Errorf("computed inner expression: %T cannot compile to SQL", expr)
}

func computedInnerBinarySql(node computed.BinaryExpr) (string, error) {
	left, err := computedInnerSqlExpr(node.Left)
	if err != nil {
		return "", err
	}
	right, err := computedInnerSqlExpr(node.Right)
	if err != nil {
		return "", err
	}
	operator, err := computedInnerSqlOperator(node.Op)
	if err != nil {
		return "", err
	}
	if node.Op == computed.OpDivide {
		// The Go evaluator divides in decimal; casting the dividend keeps SQL from
		// integer-dividing and silently disagreeing with it.
		left = "(" + left + ")::numeric"
	}
	return "(" + left + " " + operator + " " + right + ")", nil
}

func computedInnerSqlOperator(op computed.BinaryOperator) (string, error) {
	switch op {
	case computed.OpAdd:
		return "+", nil
	case computed.OpSubtract:
		return "-", nil
	case computed.OpMultiply:
		return "*", nil
	case computed.OpDivide:
		return "/", nil
	case computed.OpModulo:
		return "%", nil
	}
	return "", errors.Errorf("computed inner expression: operator %q cannot compile to SQL", op)
}

func computedInnerFunctionSql(node computed.FunctionExpr) (string, error) {
	args := make([]string, len(node.Args))
	for i, arg := range node.Args {
		compiled, err := computedInnerSqlExpr(arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}
	switch node.Name {
	case "coalesce":
		return "COALESCE(" + strings.Join(args, ", ") + ")", nil
	case "nullif":
		return "NULLIF(" + strings.Join(args, ", ") + ")", nil
	}
	return "", errors.Errorf("computed inner expression: function %q cannot compile to SQL", node.Name)
}
