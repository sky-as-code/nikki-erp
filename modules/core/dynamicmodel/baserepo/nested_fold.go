package baserepo

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// Single-statement edge projection.
//
// A query builder that implements orm.NestedProjectionCapable projects edge fields inside the
// root SELECT: to-one leaves as flat "edge.leaf" columns, to-many edges as one jsonb array
// column. The repository then has no per-row follow-up reads to issue; it scans the flat row
// and folds it back into the nested shape the API promises — the same shape the per-row
// hydration path produces, so callers cannot tell which builder served them:
//
//   - a to-one edge lands under the edge name as one map, or nil when the joined primary key is
//     NULL (the key is present either way);
//   - a to-many edge lands as a slice of maps, empty when nothing matched;
//   - destination tenant keys never reach the caller;
//   - every value is converted through the destination field's data type, as a direct read
//     would convert it.

// nestedProjectionPlan asks a capable builder for its folding plan. A nil plan with no errors
// means either the builder is not capable or the columns name no edge, and the caller falls
// back to the per-row hydration path.
func (this *BaseDynamicRepositoryImpl) nestedProjectionPlan(columns []string) (*orm.NestedProjection, ft.ClientErrors, error) {
	capable, ok := this.queryBuilder.(orm.NestedProjectionCapable)
	if !ok || len(columns) == 0 {
		return nil, nil, nil
	}
	plan, cErrs, err := capable.PlanNestedProjection(this.schema, dmodel.GetSchemaRegistry(), orm.ToSelectColumns(columns))
	if err != nil {
		return nil, nil, err
	}
	if cErrs != nil && cErrs.Count() > 0 {
		return nil, *cErrs, nil
	}
	return plan, nil, nil
}

func (this *BaseDynamicRepositoryImpl) projectsNestedEdges() bool {
	_, ok := this.queryBuilder.(orm.NestedProjectionCapable)
	return ok
}

func (this *BaseDynamicRepositoryImpl) searchWithProjectedEdges(
	ctx corectx.Context, param dyn.RepoSearchParam, plan *orm.NestedProjection,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	merged := this.injectTenantIntoGraph(ctx, param.Graph)
	columns := this.ensureSystemColumns(param.Fields)
	total, countClientErrs, err := this.countRowsMatchingGraph(ctx, merged, param.Language, columns)
	if err != nil {
		return nil, err
	}
	if len(countClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: countClientErrs}, nil
	}
	param.Fields = columns
	rows, scanClientErrs, err := this.runProjectedScan(ctx, merged, param, plan)
	if err != nil {
		return nil, err
	}
	if len(scanClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: scanClientErrs}, nil
	}
	if param.Size <= 0 {
		total = len(rows)
	}
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data: dyn.PagedResultData[dmodel.DynamicFields]{
			Items: rows, Total: total, Page: param.Page, Size: param.Size,
		},
		HasData: len(rows) != 0,
	}, nil
}

func (this *BaseDynamicRepositoryImpl) getOneWithProjectedEdges(
	ctx corectx.Context, param dyn.RepoGetOneParam, plan *orm.NestedProjection,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	graph, err := this.buildFindOneGraph(param.Filter)
	if err != nil {
		return nil, err
	}
	graph = this.injectTenantIntoGraph(ctx, graph)
	rows, scanErrs, err := this.runProjectedScan(ctx, graph, dyn.RepoSearchParam{
		Fields: this.ensureSystemColumns(param.Fields), Page: 0, Size: 1,
	}, plan)
	if err != nil {
		return nil, err
	}
	if len(scanErrs) > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: scanErrs}, nil
	}
	if len(rows) == 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{HasData: false}, nil
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: rows[0], HasData: true}, nil
}

func (this *BaseDynamicRepositoryImpl) runProjectedScan(
	ctx corectx.Context, graph *dmodel.SearchGraph, param dyn.RepoSearchParam, plan *orm.NestedProjection,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	registry := dmodel.GetSchemaRegistry()
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(this.schema, registry, graph, orm.SqlSelectGraphOpts{
		Columns:         orm.ToSelectColumns(param.Fields),
		Page:            param.Page,
		Size:            param.Size,
		Language:        param.Language,
		ComputedContext: param.ComputedContext,
	})
	if err != nil {
		return nil, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, *qbClientErrs, nil
	}
	this.logQuery(*sqlQuery)
	rows, err := this.queryAndScan(ctx, *sqlQuery, projectedScanFields(registry, this.schema, param.Fields, plan))
	if err != nil {
		return nil, nil, err
	}
	if err := foldNestedRows(registry, rows, plan); err != nil {
		return nil, nil, err
	}
	if rows == nil {
		rows = []dmodel.DynamicFields{}
	}
	return rows, nil, nil
}

// projectedScanFields maps every scanned column to the field that converts it: root columns to
// the root schema, flat to-one aliases to their destination field. To-many jsonb columns stay
// unmapped so the scanner hands the raw document to the fold.
func projectedScanFields(
	registry *dmodel.SchemaRegistry, root *dmodel.ModelSchema, columns []string, plan *orm.NestedProjection,
) map[string]*dmodel.ModelField {
	rootOnly := make([]string, 0, len(columns))
	for _, col := range columns {
		if !strings.Contains(col, ".") {
			if field, ok := root.Field(col); ok && !field.IsEdgeModel() {
				rootOnly = append(rootOnly, col)
			}
		}
	}
	out := map[string]*dmodel.ModelField{}
	for _, col := range rootOnly {
		if field, ok := root.Field(col); ok {
			out[field.Name()] = field
		}
	}
	if plan == nil {
		return out
	}
	for _, entry := range plan.ToOne {
		dest := registry.Get(entry.DestSchemaName)
		if dest == nil {
			continue
		}
		for alias, fieldName := range entry.Leaves {
			if field, ok := dest.Field(fieldName); ok {
				out[alias] = field
			}
		}
	}
	return out
}

// foldNestedRows turns the flat aliases of every row into nested maps and slices in place.
func foldNestedRows(registry *dmodel.SchemaRegistry, rows []dmodel.DynamicFields, plan *orm.NestedProjection) error {
	if plan == nil {
		return nil
	}
	toOneKeys := sortedByDepth(mapKeys(plan.ToOne))
	toManyKeys := sortedByDepth(mapKeys(plan.ToMany))
	for i := range rows {
		for _, key := range toOneKeys {
			if err := foldToOne(registry, rows[i], key, plan.ToOne[key]); err != nil {
				return err
			}
		}
		for _, key := range toManyKeys {
			if err := foldToMany(registry, rows[i], key, plan.ToMany[key]); err != nil {
				return err
			}
		}
	}
	return nil
}

func foldToOne(registry *dmodel.SchemaRegistry, row dmodel.DynamicFields, key string, entry orm.ToOneEdgeProjection) error {
	dest := registry.Get(entry.DestSchemaName)
	if dest == nil {
		return errors.Errorf("foldToOne: schema %q not in registry", entry.DestSchemaName)
	}
	present := len(entry.PkAliases) > 0
	for _, pkAlias := range entry.PkAliases {
		if row[pkAlias] == nil {
			present = false
			break
		}
	}
	var nested dmodel.DynamicFields
	if present {
		nested = make(dmodel.DynamicFields, len(entry.Leaves))
		tenantKey := dest.TenantKey()
		for alias, fieldName := range entry.Leaves {
			value, ok := row[alias]
			if !ok || fieldName == tenantKey {
				continue
			}
			nested[fieldName] = value
		}
	}
	for alias := range entry.Leaves {
		delete(row, alias)
	}
	return setNested(row, key, nested)
}

func foldToMany(registry *dmodel.SchemaRegistry, row dmodel.DynamicFields, key string, entry orm.ToManyEdgeProjection) error {
	dest := registry.Get(entry.DestSchemaName)
	if dest == nil {
		return errors.Errorf("foldToMany: schema %q not in registry", entry.DestSchemaName)
	}
	raw, ok := row[entry.ColumnAlias]
	delete(row, entry.ColumnAlias)
	items := []dmodel.DynamicFields{}
	if ok && raw != nil {
		decoded, err := decodeJsonbRows(raw)
		if err != nil {
			return errors.Wrapf(err, "foldToMany: edge %q", key)
		}
		tenantKey := dest.TenantKey()
		for _, element := range decoded {
			item := make(dmodel.DynamicFields, len(element))
			for fieldName, value := range element {
				if fieldName == tenantKey || value == nil {
					continue
				}
				converted, err := convertJsonValue(dest, fieldName, value)
				if err != nil {
					return errors.Wrapf(err, "foldToMany: edge %q field %q", key, fieldName)
				}
				if converted != nil {
					item[fieldName] = converted
				}
			}
			items = append(items, item)
		}
	}
	return setNested(row, key, items)
}

func decodeJsonbRows(raw any) ([]map[string]any, error) {
	var data []byte
	switch typed := raw.(type) {
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		return nil, errors.Errorf("expected jsonb text, got %T", raw)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var out []map[string]any
	if err := decoder.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// convertJsonValue converts one jsonb element value through the destination field's data type.
// Numbers arrive as json.Number and are handed over as their exact decimal text, which every
// numeric type parses without going through float64.
func convertJsonValue(dest *dmodel.ModelSchema, fieldName string, value any) (any, error) {
	field, ok := dest.Field(fieldName)
	if !ok {
		return value, nil
	}
	if number, isNumber := value.(json.Number); isNumber {
		value = number.String()
	}
	converted, err := field.DataType().TryConvert(value, field.DataType().Options())
	if err != nil {
		return nil, err
	}
	if converted.Get() == nil {
		return nil, nil
	}
	return *converted.Get(), nil
}

// setNested stores value at the dotted key, creating no intermediate map: when a parent edge is
// absent (nil) the child has nowhere to live and is dropped, matching a per-row hydration that
// could not have followed the missing relation.
func setNested(row dmodel.DynamicFields, key string, value any) error {
	segments := strings.Split(key, ".")
	current := row
	for _, seg := range segments[:len(segments)-1] {
		child, ok := current[seg]
		if !ok || child == nil {
			return nil
		}
		next, isMap := child.(dmodel.DynamicFields)
		if !isMap {
			return errors.Errorf("setNested: %q is not a nested record on the row", seg)
		}
		current = next
	}
	last := segments[len(segments)-1]
	switch typed := value.(type) {
	case dmodel.DynamicFields:
		if typed == nil {
			current[last] = nil
			return nil
		}
		current[last] = typed
	default:
		current[last] = value
	}
	return nil
}

func mapKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}

func sortedByDepth(keys []string) []string {
	sort.SliceStable(keys, func(i, j int) bool {
		di, dj := strings.Count(keys[i], "."), strings.Count(keys[j], ".")
		if di != dj {
			return di < dj
		}
		return keys[i] < keys[j]
	})
	return keys
}
