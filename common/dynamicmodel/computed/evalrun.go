package computed

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// SourceSearchFn fetches source rows for a batched related read: the rows of schemaName whose
// keyColumn is IN keys, projected down to fields. The caller supplies it so this package stays
// free of repository/engine dependencies (and a test can stub it).
type SourceSearchFn func(schemaName string, keyColumn string, keys []any, fields []string) ([]dmodel.DynamicFields, error)

// FunctionInvokerFn runs a registered computed-field function over a whole page, returning one
// value per row in the order given. Like SourceSearchFn it is supplied by the caller: the function
// registry lives on the engine, and the request context it needs is captured in the closure there,
// so this package never learns about either.
type FunctionInvokerFn func(functionName string, fieldName string, rows []dmodel.DynamicFields) ([]any, error)

// EvalDeps carries what evaluation cannot do for itself. Grouped into a struct so a later stage
// needing a third capability does not churn every Apply call site.
type EvalDeps struct {
	Search SourceSearchFn
	Invoke FunctionInvokerFn
}

// Apply evaluates the plan over one page of rows, in place: batched related fills first, then
// function calls, then expression fields in dependency order. Functions run before expressions
// because an expression may read a function field, never the other way round — a function is a
// whole-field root and can never sit inside an expression tree.
//
// A row whose source record is missing keeps its related fields absent — never zero values —
// matching the virtual-field convention this generalizes.
func (this *EvalPlan) Apply(rows []dmodel.DynamicFields, deps EvalDeps) error {
	if len(rows) == 0 {
		return nil
	}
	for _, read := range this.RelatedReads {
		if err := this.applyRelatedRead(read, rows, deps.Search); err != nil {
			return err
		}
	}
	for _, call := range this.FunctionCalls {
		if err := this.applyFunctionCall(call, rows, deps.Invoke); err != nil {
			return err
		}
	}
	return this.applyExpressions(rows)
}

// applyFunctionCall runs one registered function over the page and assigns its results.
//
// A short result slice would silently shift values onto the wrong rows, so a length mismatch is an
// error rather than a best-effort assignment.
func (this *EvalPlan) applyFunctionCall(
	call FunctionCall, rows []dmodel.DynamicFields, invoke FunctionInvokerFn,
) error {
	if invoke == nil {
		return errors.Errorf(
			"computed field %s.%s needs function %q but no function invoker was supplied",
			this.SchemaName, call.Fields[0], call.FunctionName)
	}
	for _, fieldName := range call.Fields {
		values, err := invoke(call.FunctionName, fieldName, rows)
		if err != nil {
			return errors.Wrapf(err, "computed field %s.%s via function %q",
				this.SchemaName, fieldName, call.FunctionName)
		}
		if len(values) != len(rows) {
			return errors.Errorf(
				"computed function %q returned %d values for %d rows of %s.%s",
				call.FunctionName, len(values), len(rows), this.SchemaName, fieldName)
		}
		for i, row := range rows {
			row[fieldName] = values[i]
		}
	}
	return nil
}

func (this *EvalPlan) applyRelatedRead(
	read RelatedRead, rows []dmodel.DynamicFields, search SourceSearchFn,
) error {
	keys := distinctKeys(rows, read.FkColumn)
	if len(keys) == 0 {
		return nil
	}
	projection := make([]string, 0, len(read.Leaves)+1)
	projection = append(projection, read.RefColumn)
	for _, leaf := range read.Leaves {
		if leaf != read.RefColumn {
			projection = append(projection, leaf)
		}
	}

	sourceRows, err := search(read.SchemaName, read.RefColumn, keys, projection)
	if err != nil {
		return errors.Wrapf(err, "batched read of %q for computed fields", read.SchemaName)
	}
	byKey := indexByColumn(sourceRows, read.RefColumn)
	copyLeaves(read, rows, byKey)
	return nil
}

func copyLeaves(read RelatedRead, rows []dmodel.DynamicFields, byKey map[string]dmodel.DynamicFields) {
	for _, row := range rows {
		key := normalizeValue(row[read.FkColumn])
		if key == nil {
			continue
		}
		source, ok := byKey[keyString(key)]
		if !ok {
			// A missing source must read as "unknown", not as a record with empty values.
			continue
		}
		for fieldName, leaf := range read.Leaves {
			row[fieldName] = source[leaf]
		}
	}
}

func (this *EvalPlan) applyExpressions(rows []dmodel.DynamicFields) error {
	for _, name := range this.Wanted {
		fieldPlan := this.schemaPlan.Fields[name]
		if fieldPlan.Def.Kind != ComputeExpression {
			continue
		}
		for _, row := range rows {
			value, err := Eval(fieldPlan.Def.Expression, row)
			if err != nil {
				return errors.Wrapf(err, "computed field %s.%s", this.SchemaName, name)
			}
			row[name] = value
		}
	}
	return nil
}

func distinctKeys(rows []dmodel.DynamicFields, column string) []any {
	seen := map[string]bool{}
	var keys []any
	for _, row := range rows {
		value := normalizeValue(row[column])
		if value == nil {
			continue
		}
		text := keyString(value)
		if !seen[text] {
			seen[text] = true
			keys = append(keys, value)
		}
	}
	return keys
}

func indexByColumn(rows []dmodel.DynamicFields, column string) map[string]dmodel.DynamicFields {
	index := make(map[string]dmodel.DynamicFields, len(rows))
	for _, row := range rows {
		value := normalizeValue(row[column])
		if value != nil {
			index[keyString(value)] = row
		}
	}
	return index
}

func keyString(value any) string {
	return fmt.Sprint(value)
}
