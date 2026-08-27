package services

import (
	"fmt"

	"github.com/shopspring/decimal"

	ft "github.com/sky-as-code/nikki-erp/common/fault"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// fakeRepo is an in-memory TaxSearcher that really evaluates the search graph.
//
// It matters that it evaluates rather than ignores the graph: three of the bugs found in this
// module were about which rows a lookup actually selects, and a fake that returns everything it
// holds regardless of the filter would have passed every one of them. The graph's accessors are
// exported, so the evaluation here is the real predicate the lookup built — only the storage is
// substituted.
//
// The engine's own SQL translation is not reproduced, and this is not a substitute for exercising
// the real repository. What it buys is that the lookups, the resolver and the config loader can be
// tested over configurations too awkward to seed: ambiguous versions, cyclic components, a rate
// that expires mid-year.
type fakeRepo struct {
	rows []dmodel.DynamicFields

	// calls records every graph the repository was asked to evaluate, so a test can assert that a
	// lookup filtered on what it claimed to.
	calls int
}

func newFakeRepo(rows ...dmodel.DynamicFields) *fakeRepo {
	return &fakeRepo{rows: rows}
}

func (this *fakeRepo) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.calls++

	matched := make([]dmodel.DynamicFields, 0, len(this.rows))
	for _, row := range this.rows {
		if graphMatches(param.Graph, row) {
			matched = append(matched, row)
		}
	}
	// The lookups pass a limit as the page size, and a lookup that reads fewer rows than exist is
	// a real failure mode — FindEffectiveDefinitionVersion counting matches depends on seeing them
	// all — so the fake honours it rather than quietly returning everything.
	if param.Size > 0 && len(matched) > param.Size {
		matched = matched[:param.Size]
	}

	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		HasData: true,
		Data: dyn.PagedResultData[dmodel.DynamicFields]{
			Items: matched,
			Total: len(matched),
			Page:  param.Page,
			Size:  param.Size,
		},
	}, nil
}

// graphMatches evaluates a search graph against one row.
//
// Only the shapes the tax lookups actually build are supported — a bare condition, and an AND of
// conditions. An unsupported shape panics rather than silently matching, because a fake that
// quietly answers "yes" to a filter it does not understand is how a test comes to prove nothing.
func graphMatches(graph *dmodel.SearchGraph, row dmodel.DynamicFields) bool {
	if graph == nil {
		return true
	}
	if condition := graph.GetCondition(); len(condition) > 0 {
		return conditionMatches(condition, row)
	}

	nodes := graph.GetAnd()
	if len(nodes) == 0 {
		if len(graph.GetOr()) > 0 {
			panic("fakeRepo: OR graphs are not supported; no tax lookup builds one")
		}
		return true
	}
	for _, node := range nodes {
		if !nodeMatches(node, row) {
			return false
		}
	}
	return true
}

func nodeMatches(node dmodel.SearchNode, row dmodel.DynamicFields) bool {
	if condition := node.GetCondition(); len(condition) > 0 {
		return conditionMatches(condition, row)
	}
	for _, child := range node.GetAnd() {
		if !nodeMatches(child, row) {
			return false
		}
	}
	return true
}

func conditionMatches(condition dmodel.Condition, row dmodel.DynamicFields) bool {
	field := condition.Field()
	operator := condition.Operator()
	if operator != dmodel.Equals {
		panic(fmt.Sprintf("fakeRepo: unsupported operator %q on field %q", operator, field))
	}

	stored, present := row[field]
	if !present {
		return false
	}
	return fmt.Sprint(stored) == fmt.Sprint(condition.Value())
}

// decimalOf is a terse decimal literal for tests, where an unparseable constant is a typo in the
// test rather than a condition worth handling.
func decimalOf(value string) decimal.Decimal {
	return decimal.RequireFromString(value)
}

// dateOf builds a ModelDate for a test row.
//
// The dynamic-field getters convert only from ModelDate or time.Time — a bare "2025-01-01" string
// reads back as nil, which would make an effective-period test silently pass for the wrong reason.
func dateOf(value string) model.ModelDate {
	parsed, err := model.ParseModelDate(value)
	if err != nil {
		panic("dateOf: " + value + " is not a calendar date: " + err.Error())
	}
	return parsed
}

// clientErrorsWith builds the rejection shape an application service returns for a business
// violation, so a fake external service can refuse the way a real one does.
func clientErrorsWith(key, message string) *ft.ClientErrors {
	cErrs := ft.NewClientErrors()
	cErrs.Append(*ft.NewBusinessViolation("uom", key, message))
	return cErrs
}
