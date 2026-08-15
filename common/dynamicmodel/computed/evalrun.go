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

// Apply evaluates the plan over one page of rows, in place: batched related fills first, then
// expression fields in dependency order. A row whose source record is missing keeps its related
// fields absent — never zero values — matching the virtual-field convention this generalizes.
func (this *EvalPlan) Apply(rows []dmodel.DynamicFields, search SourceSearchFn) error {
	if len(rows) == 0 {
		return nil
	}
	for _, read := range this.RelatedReads {
		if err := this.applyRelatedRead(read, rows, search); err != nil {
			return err
		}
	}
	return this.applyExpressions(rows)
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
