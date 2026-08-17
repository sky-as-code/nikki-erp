package orm

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/huandu/go-sqlbuilder"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"go.bryk.io/pkg/errors"
)

const selectDistinctToken = "DISTINCT::"

// AsDistinct returns a column token that triggers SELECT DISTINCT for the whole graph query.
func (this SelectColumn) AsDistinct() SelectColumn {
	raw := strings.TrimSpace(string(this))
	if strings.HasPrefix(strings.ToLower(raw), strings.ToLower(selectDistinctToken)) {
		return this
	}
	return SelectColumn(selectDistinctToken + raw)
}

func (this SelectColumn) rawString() string {
	return string(this)
}

// ToSelectColumns converts API column strings into SelectColumn values.
func ToSelectColumns(cols []string) []SelectColumn {
	if len(cols) == 0 {
		return nil
	}
	out := make([]SelectColumn, len(cols))
	for i, c := range cols {
		out[i] = SelectColumn(c)
	}
	return out
}

func selectColumnHasDistinctPrefix(col SelectColumn) bool {
	raw := strings.TrimSpace(col.rawString())
	return strings.HasPrefix(strings.ToLower(raw), strings.ToLower(selectDistinctToken))
}

func selectColumnStripDistinctPrefix(col SelectColumn) string {
	raw := strings.TrimSpace(col.rawString())
	lower := strings.ToLower(raw)
	const p = "distinct::"
	if strings.HasPrefix(lower, p) {
		return strings.TrimSpace(raw[len(p):])
	}
	return raw
}

// joinPlanningPath returns a dotted path used for graph join discovery, or empty when the token is not a plain field path.
func (this SelectColumn) joinPlanningPath() string {
	raw := strings.TrimSpace(this.rawString())
	if raw == "" {
		return ""
	}
	inner := selectColumnStripDistinctPrefix(this)
	if strings.Contains(inner, "(") {
		return ""
	}
	return inner
}

func anySelectColumnDistinct(cols []SelectColumn) bool {
	for _, c := range cols {
		if selectColumnHasDistinctPrefix(c) {
			return true
		}
	}
	return false
}

func parseAllowedAggregate(col SelectColumn) (funcUpper string, innerWithPossibleDistinct string, ok bool) {
	compact := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, col.rawString())
	if compact == "" {
		return "", "", false
	}
	open := strings.IndexByte(compact, '(')
	close := strings.LastIndexByte(compact, ')')
	if open <= 0 || close <= open || close != len(compact)-1 {
		return "", "", false
	}
	fn := strings.ToUpper(compact[:open])
	inner := compact[open+1 : close]
	if inner == "" {
		return "", "", false
	}
	switch fn {
	case "COUNT", "MAX", "MIN", "AVG", "SUM":
		return fn, inner, true
	default:
		return "", "", false
	}
}

func clientErrorsInvalidSelectAggregate(token string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(token, ft.ErrorKey("err_invalid_select_aggregate"),
			"select column must be a field path, DISTINCT::field, or COUNT|MAX|MIN|AVG|SUM(field)"),
	}
}

func buildAggregateSelectExpr(
	planner *joinPlanner, token SelectColumn, funcUpper, inner string,
) (string, ft.ClientErrors, error) {
	innerCol := SelectColumn(inner)
	distinctArg := selectColumnHasDistinctPrefix(innerCol)
	innerPath := selectColumnStripDistinctPrefix(innerCol)
	expr, err := planner.selectExprForColumn(innerPath)
	if err != nil {
		return "", nil, errors.Wrap(err, "buildAggregateSelectExpr")
	}
	if distinctArg && funcUpper != "COUNT" {
		return "", clientErrorsInvalidSelectAggregate(token.rawString()), nil
	}
	var out string
	if distinctArg {
		out = funcUpper + "(DISTINCT " + expr + ")"
	} else {
		out = funcUpper + "(" + expr + ")"
	}
	return out, nil, nil
}

// orderRefsForDistinct returns the ORDER BY refs that must be forced into the select list.
//
// Under SELECT DISTINCT, PostgreSQL rejects an ORDER BY expression that is not projected. The
// refs are already-built SQL ("t1.\"label\" ASC"), so the direction suffix is trimmed off.
func orderRefsForDistinct(isDistinct bool, orderExprs []string) []string {
	if !isDistinct {
		return nil
	}
	refs := make([]string, 0, len(orderExprs))
	for _, expr := range orderExprs {
		if ref := strings.TrimSpace(strings.TrimSuffix(
			strings.TrimSuffix(strings.TrimSpace(expr), " DESC"), " ASC")); ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}

// applySelectColumns writes the projection. ctxValues binds "${ctx.key}" references of
// SQL-computed fields; extraRefs are already-resolved SQL refs appended when they are not
// already projected; see orderRefsForDistinct.
func (this *PgQueryBuilder) applySelectColumns(
	sb *sqlbuilder.SelectBuilder, planner *joinPlanner, columns []SelectColumn,
	ctxValues map[string]any, extraRefs ...string,
) error {
	if len(columns) == 0 {
		wildcard := "*"
		if planner != nil && planner.usesJoins() {
			planner.ensureRootAliased()
			wildcard = fmt.Sprintf("%s.*", planner.rootAlias)
		}
		// The wildcard covers the root table only, so an ORDER BY on a joined column is still
		// absent from the projection — which SELECT DISTINCT rejects outright. extraRefs carries
		// exactly those sort expressions, so they are appended here too.
		sb.Select(append([]string{wildcard}, extraRefs...)...)
		return nil
	}
	selectCols := make([]string, 0, len(columns))
	sqlComputedCount := 0
	for _, col := range columns {
		if fn, inner, ok := parseAllowedAggregate(col); ok {
			expr, cErrs, err := buildAggregateSelectExpr(planner, col, fn, inner)
			if err != nil {
				return err
			}
			if len(cErrs) > 0 {
				return wrapClientSqlErrors(cErrs)
			}
			selectCols = append(selectCols, expr)
			continue
		}
		path := selectColumnStripDistinctPrefix(col)
		if strings.Contains(path, "(") {
			return wrapClientSqlErrors(clientErrorsInvalidSelectAggregate(col.rawString()))
		}
		// A virtual scalar has no column to project. An SQL-computed one (aggregate/exists/
		// lookup) materializes as a correlated subquery right here; any other virtual scalar
		// stays a legal request that a service fills after the read, so it is skipped.
		if planner != nil && planner.isVirtualScalarPath(path) {
			expr, emitted, err := this.computedSelectExpr(planner, path, ctxValues, &sqlComputedCount)
			if err != nil {
				return err
			}
			if emitted {
				selectCols = append(selectCols, expr)
			}
			continue
		}
		expr, err := planner.selectExprForColumn(path)
		if err != nil {
			return errors.Wrap(err, "applySelectColumns")
		}
		selectCols = append(selectCols, expr)
	}
	// Every requested column was virtual. An empty list would make sqlbuilder emit "SELECT ",
	// and falling through to "SELECT *" would contradict the caller's projection, so anchor the
	// query on the primary keys instead.
	if len(selectCols) == 0 && planner != nil {
		selectCols = planner.primaryKeySelectRefs()
	}
	for _, ref := range extraRefs {
		if !containsString(selectCols, ref) {
			selectCols = append(selectCols, ref)
		}
	}
	sb.Select(selectCols...)
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
