package advanced

import (
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// applySelectColumns writes the projection. extraRefs are already-rendered SQL refs that must
// be present under SELECT DISTINCT (the ORDER BY expressions); they are appended when missing.
func (this *queryPlan) applySelectColumns(sb *sqlbuilder.SelectBuilder, columns []orm.SelectColumn, extraRefs ...string) error {
	if len(columns) == 0 {
		sb.Select(append([]string{this.rootWildcard()}, extraRefs...)...)
		return nil
	}
	proj := newProjection()
	allColumns := false
	for _, col := range columns {
		if col == orm.SelectAllColumns {
			allColumns = true
			continue
		}
		if fn, inner, ok := orm.ParseAllowedAggregate(col); ok {
			expr, err := this.aggregateSelectExpr(col, fn, inner)
			if err != nil {
				return err
			}
			proj.addExpr(expr)
			continue
		}
		path := orm.SelectColumnPath(col)
		if strings.Contains(path, "(") {
			return orm.WrapClientErrors(orm.ClientErrorsInvalidSelectAggregate(col.Raw()))
		}
		if err := this.projectPath(proj, path); err != nil {
			return errors.Wrap(err, "applySelectColumns")
		}
	}
	selectCols := make([]string, 0, len(proj.items)+1)
	if allColumns {
		selectCols = append(selectCols, this.rootWildcard())
	}
	for _, item := range proj.items {
		if allColumns && this.isRootColumnRef(item.expr) {
			continue
		}
		if item.groupKey == "" {
			selectCols = append(selectCols, item.expr)
			continue
		}
		expr, err := this.renderToMany(proj.groups[item.groupKey])
		if err != nil {
			return errors.Wrap(err, "applySelectColumns")
		}
		selectCols = append(selectCols, expr)
	}
	// Every requested column was virtual and Go-filled: anchor the query on the primary keys
	// rather than emitting an empty list or contradicting the caller with "*".
	if len(selectCols) == 0 {
		for _, key := range this.root.PrimaryKeys() {
			selectCols = append(selectCols, this.columnRef("", key))
		}
	}
	for _, ref := range extraRefs {
		if allColumns && this.isRootColumnRef(ref) {
			continue
		}
		if !containsString(selectCols, ref) {
			selectCols = append(selectCols, ref)
		}
	}
	sb.Select(selectCols...)
	return nil
}

func (this *queryPlan) rootWildcard() string {
	if alias := this.join.RootAlias(); alias != "" {
		return alias + ".*"
	}
	return "*"
}

// isRootColumnRef tells whether expr is a bare physical column of the root, which `*` already
// projects; keeping it would return the column twice. Primary keys registered by edge hops and
// operands appended for computed fields both arrive this way.
func (this *queryPlan) isRootColumnRef(expr string) bool {
	for _, field := range this.root.Columns() {
		if expr == this.columnRef("", field.Name()) {
			return true
		}
	}
	return false
}

// aggregateSelectExpr renders a COUNT|MAX|MIN|AVG|SUM(field) token; COUNT(DISTINCT::field) is
// the only distinct form allowed.
func (this *queryPlan) aggregateSelectExpr(token orm.SelectColumn, funcUpper, inner string) (string, error) {
	innerCol := orm.SelectColumn(inner)
	distinctArg := orm.SelectColumnIsDistinct(innerCol)
	innerPath := orm.SelectColumnPath(innerCol)
	if distinctArg && funcUpper != "COUNT" {
		return "", orm.WrapClientErrors(orm.ClientErrorsInvalidSelectAggregate(token.Raw()))
	}
	ref, err := this.resolve(innerPath, useSelect)
	if err != nil {
		return "", errors.Wrap(err, "aggregateSelectExpr")
	}
	if ref.kind == refComputedGo {
		return "", orm.WrapClientErrors(orm.ClientErrorsVirtualFieldUnavailable(innerPath))
	}
	if distinctArg {
		return funcUpper + "(DISTINCT " + ref.sql + ")", nil
	}
	return funcUpper + "(" + ref.sql + ")", nil
}

func sqlComputedLimit() int {
	return computed.ActiveLimits().MaxSqlComputedFieldsPerRequest
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
