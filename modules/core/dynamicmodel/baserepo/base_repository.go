package baserepo

import (
	"reflect"
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type NewBaseRepositoryParam struct {
	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	Logger       logging.LoggerService
	QueryBuilder orm.QueryBuilder
	Schema       *dmodel.ModelSchema
}

func NewBaseRepositoryImpl(param NewBaseRepositoryParam) coredyn.BaseRepository {
	sqlDebugEnabled := param.ConfigSvc.GetBool(c.DbDebugEnabled)

	return &BaseRepositoryImpl{
		client:          param.Client,
		queryBuilder:    param.QueryBuilder,
		logger:          param.Logger,
		schema:          param.Schema,
		sqlDebugEnabled: sqlDebugEnabled,
	}
}

// Ensure interface implementation at compile time.
var _ coredyn.BaseRepository = (*BaseRepositoryImpl)(nil)

type BaseRepositoryImpl struct {
	client          orm.DbClient
	queryBuilder    orm.QueryBuilder
	logger          logging.LoggerService
	schema          *dmodel.ModelSchema
	sqlDebugEnabled bool
}

func (this *BaseRepositoryImpl) GetSchema() *dmodel.ModelSchema {
	return this.schema
}

// Insert inserts a record to database. If the schema defines "created_at", sets current UTC timestamp.
// On success, Data holds the inserted field map; IsEmpty is false when Data is non-nil.
func (this *BaseRepositoryImpl) Insert(ctx corectx.Context, data dmodel.DynamicFields) (
	*crud.OpResult[dmodel.DynamicFields], error,
) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	this.trySetCreatedAt(data)
	this.trySetEtag(data)
	qbRes, err := this.queryBuilder.SqlInsert(this.schema, data)
	if err != nil {
		return nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: qbRes.ClientErrors}, nil
	}
	query := qbRes.Data

	this.logQuery(query)
	_, err = this.client.Exec(ctx, query)
	if err != nil {
		return nil, err
	}
	return &crud.OpResult[dmodel.DynamicFields]{Data: data, IsEmpty: false}, nil
}

// Update updates a record. The data map must contain primary keys and tenant key.
// If the schema defines "updated_at", sets current UTC timestamp.
func (this *BaseRepositoryImpl) Update(ctx corectx.Context, data dmodel.DynamicFields) (
	*crud.OpResult[dmodel.DynamicFields], error,
) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	filters := this.extractKeyMap(data)
	this.trySetUpdatedAt(data)
	prevEtag := this.trySetEtag(data)
	if len(prevEtag) > 0 {
		filters[basemodel.FieldEtag] = prevEtag
	}
	qbRes, err := this.queryBuilder.SqlUpdateEqual(this.schema, data, filters)
	if err != nil {
		return nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: qbRes.ClientErrors}, nil
	}
	query := qbRes.Data

	this.logQuery(query)
	_, err = this.client.Exec(ctx, query)
	if err != nil {
		return nil, err
	}
	return &crud.OpResult[dmodel.DynamicFields]{Data: data, IsEmpty: false}, nil
}

// Implements BaseRepository interface
func (this *BaseRepositoryImpl) GetOne(ctx corectx.Context, param coredyn.GetOneParam) (
	*crud.OpResult[dmodel.DynamicFields], error,
) {
	if vErr := this.validateGetOneColumnsAndFilter(param.Columns, param.Filter); vErr != nil {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: ft.ClientErrors{*vErr}}, nil
	}
	if err := this.ensureTenantKeyIn(param.Filter); err != nil {
		return nil, err
	}
	graph, err := this.buildFindOneGraph(param.Filter, param.IncludeArchived)
	if err != nil {
		return nil, err
	}
	qbRes, err := this.queryBuilder.SqlSelectGraph(this.schema, graph, orm.SqlSelectGraphOpts{
		Columns: param.Columns,
	})
	if err != nil {
		return nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: qbRes.ClientErrors}, nil
	}
	query := qbRes.Data

	this.logQuery(query)
	rows, err := this.queryAndScan(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &crud.OpResult[dmodel.DynamicFields]{IsEmpty: true}, nil
	}
	return &crud.OpResult[dmodel.DynamicFields]{Data: rows[0], IsEmpty: false}, nil
}

// Search fetches records matching the SearchGraph criteria.
// When the schema is tenant-scoped, filter must be provided and contain the tenant key.
// Data uses PagedResult: Total is from COUNT when Size > 0, otherwise len(Items).
func (this *BaseRepositoryImpl) Search(ctx corectx.Context, param coredyn.SearchParam) (
	*crud.OpResult[crud.PagedResultData[dmodel.DynamicFields]], error,
) {
	// merged, err := this.mergeFilterIntoGraph(param.Graph, param.Filter)
	// if err != nil {
	// 	return nil, err
	// }
	merged := param.Graph
	page := param.Page
	size := param.Size
	var total int
	total, countClientErrs, err := this.countRowsMatchingGraph(ctx, merged)
	if err != nil {
		return nil, err
	}
	if len(countClientErrs) > 0 {
		return &crud.OpResult[crud.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: countClientErrs,
		}, nil
	}
	rows, scanClientErrs, err := this.runSelectGraphScan(ctx, merged, param)
	if err != nil {
		return nil, err
	}
	if len(scanClientErrs) > 0 {
		return &crud.OpResult[crud.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: scanClientErrs,
		}, nil
	}
	if size <= 0 {
		total = len(rows)
	}
	paged := crud.PagedResultData[dmodel.DynamicFields]{
		Items: rows,
		Total: total,
		Page:  page,
		Size:  size,
	}
	return &crud.OpResult[crud.PagedResultData[dmodel.DynamicFields]]{
		Data:    paged,
		IsEmpty: len(rows) == 0,
	}, nil
}

func (this *BaseRepositoryImpl) countRowsMatchingGraph(
	ctx corectx.Context, graph *dmodel.SearchGraph,
) (int, ft.ClientErrors, error) {
	qbRes, err := this.queryBuilder.SqlCountGraph(this.schema, graph)
	if err != nil {
		return 0, nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return 0, qbRes.ClientErrors, nil
	}
	countSql := qbRes.Data
	this.logQuery(countSql)
	n, ierr := this.queryScalarInt(ctx, countSql)
	return n, nil, ierr
}

func (this *BaseRepositoryImpl) runSelectGraphScan(
	ctx corectx.Context, graph *dmodel.SearchGraph, param coredyn.SearchParam,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	qbRes, err := this.queryBuilder.SqlSelectGraph(this.schema, graph, orm.SqlSelectGraphOpts{
		Columns: param.Columns,
		Page:    param.Page,
		Size:    param.Size,
	})
	if err != nil {
		return nil, nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return nil, qbRes.ClientErrors, nil
	}
	query := qbRes.Data
	this.logQuery(query)
	rows, err := this.queryAndScan(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	if rows == nil {
		return []dmodel.DynamicFields{}, nil, nil
	}
	return rows, nil, nil
}

func (this *BaseRepositoryImpl) queryScalarInt(ctx corectx.Context, query string) (int, error) {
	row := this.client.QueryRow(ctx, query)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// Archive sets archive_at to current UTC timestamp for the record identified by keys.
// Returns error if the schema does not define "archive_at" column.
func (this *BaseRepositoryImpl) Archive(ctx corectx.Context, keys dmodel.DynamicFields) (
	*crud.OpResult[dmodel.DynamicFields], error,
) {
	if _, ok := this.schema.Column(basemodel.FieldArchivedAt); !ok {
		return nil, errors.Errorf(
			"Archive: schema '%s' does not define column '%s'", this.schema.Name(), basemodel.FieldArchivedAt)
	}
	oneRes, err := this.GetOne(ctx, coredyn.GetOneParam{Filter: keys})
	if err != nil {
		return nil, err
	}
	if len(oneRes.ClientErrors) > 0 {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: oneRes.ClientErrors}, nil
	}
	if oneRes.IsEmpty {
		return &crud.OpResult[dmodel.DynamicFields]{IsEmpty: true}, nil
	}
	record := oneRes.Data
	ok := this.trySetArchivedAt(record)
	if !ok {
		return nil, errors.Errorf(
			"Archive: schema '%s' does not define column '%s'", this.schema.Name(), basemodel.FieldArchivedAt)
	}
	updRes, err := this.Update(ctx, record)
	if err != nil {
		return nil, err
	}
	if len(updRes.ClientErrors) > 0 {
		return &crud.OpResult[dmodel.DynamicFields]{ClientErrors: updRes.ClientErrors}, nil
	}
	return &crud.OpResult[dmodel.DynamicFields]{Data: updRes.Data, IsEmpty: false}, nil
}

// Delete removes the record identified by primary keys and tenant key.
// Data holds RowsAffected; IsEmpty is true when no row was deleted.
func (this *BaseRepositoryImpl) Delete(ctx corectx.Context, keys dmodel.DynamicFields) (*crud.OpResult[int], error) {
	if err := this.validateKeyMap(keys); err != nil {
		return nil, err
	}
	if err := this.ensureTenantKeyIn(keys); err != nil {
		return nil, err
	}
	qbRes, err := this.queryBuilder.SqlDeleteEqual(this.schema, keys)
	if err != nil {
		return nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return &crud.OpResult[int]{ClientErrors: qbRes.ClientErrors}, nil
	}
	query := qbRes.Data

	this.logQuery(query)
	result, err := this.client.Exec(ctx, query)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &crud.OpResult[int]{Data: int(n), IsEmpty: n == 0}, nil
}

// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
func (this *BaseRepositoryImpl) CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (
	*crud.OpResult[[][]string], error,
) {
	uniqueKeysToCheck := this.filterUniqueKeysWithValues(data)
	if len(uniqueKeysToCheck) == 0 {
		return &crud.OpResult[[][]string]{Data: nil, IsEmpty: true}, nil
	}

	qbRes, err := this.queryBuilder.SqlCheckUniqueCollisions(this.schema, uniqueKeysToCheck, data)
	if err != nil {
		return nil, err
	}
	if len(qbRes.ClientErrors) > 0 {
		return &crud.OpResult[[][]string]{ClientErrors: qbRes.ClientErrors}, nil
	}
	query := qbRes.Data.Sql
	args := qbRes.Data.Args

	this.logQuery(query)
	rows, err := this.client.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collidingKeys [][]string
	idx := 0
	for rows.Next() {
		var val int
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		if val == 1 && idx < len(uniqueKeysToCheck) {
			collidingKeys = append(collidingKeys, uniqueKeysToCheck[idx])
		}
		idx++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &crud.OpResult[[][]string]{Data: collidingKeys, IsEmpty: len(collidingKeys) == 0}, nil
}

func (this *BaseRepositoryImpl) logQuery(query string) {
	if this.sqlDebugEnabled {
		this.logger.Debug(query, nil)
	}
}

func (this *BaseRepositoryImpl) filterUniqueKeysWithValues(data dmodel.DynamicFields) [][]string {
	var result [][]string
	for _, uniqueFields := range this.schema.AllUniques() {
		if len(uniqueFields) == 0 {
			continue
		}
		hasAll := true
		for _, f := range uniqueFields {
			if _, ok := data[f]; !ok || data[f] == nil {
				hasAll = false
				break
			}
		}
		if this.schema.TenantKey() != "" {
			if _, ok := data[this.schema.TenantKey()]; !ok || data[this.schema.TenantKey()] == nil {
				hasAll = false
			}
		}
		if hasAll {
			result = append(result, uniqueFields)
		}
	}
	return result
}

func (this *BaseRepositoryImpl) trySetArchivedAt(data dmodel.DynamicFields) bool {
	if _, ok := this.schema.Column(basemodel.FieldArchivedAt); !ok {
		return false
	}
	field := this.schema.MustField(basemodel.FieldArchivedAt)
	data[basemodel.FieldArchivedAt] = *field.DataType().DefaultValue().Get()
	return true
}

func (this *BaseRepositoryImpl) trySetCreatedAt(data dmodel.DynamicFields) {
	if _, ok := this.schema.Column(basemodel.FieldCreatedAt); !ok {
		return
	}
	field := this.schema.MustField(basemodel.FieldCreatedAt)
	data[basemodel.FieldCreatedAt] = *field.DataType().DefaultValue().Get()
}

func (this *BaseRepositoryImpl) trySetUpdatedAt(data dmodel.DynamicFields) {
	if _, ok := this.schema.Column(basemodel.FieldUpdatedAt); !ok {
		return
	}
	field := this.schema.MustField(basemodel.FieldUpdatedAt)
	data[basemodel.FieldUpdatedAt] = *field.DataType().DefaultValue().Get()
}

func (this *BaseRepositoryImpl) trySetEtag(data dmodel.DynamicFields) (prevEtag string) {
	if _, ok := this.schema.Column(basemodel.FieldEtag); !ok {
		return ""
	}
	field, ok := this.schema.Field(basemodel.FieldEtag)
	if !ok {
		return ""
	}
	et := data[basemodel.FieldEtag]
	if et != nil {
		prevEtag = et.(string)
	}
	data[basemodel.FieldEtag] = *field.DataType().DefaultValue().Get()
	return
}

func (this *BaseRepositoryImpl) extractKeyMap(data dmodel.DynamicFields) dmodel.DynamicFields {
	if data == nil {
		return nil
	}
	result := make(dmodel.DynamicFields)
	for _, key := range this.schema.KeyColumns() {
		if v, ok := data[key]; ok {
			result[key] = v
		}
	}
	return result
}

func (this *BaseRepositoryImpl) validateKeyMap(keys dmodel.DynamicFields) error {
	if len(keys) == 0 {
		return errors.New("validateKeyMap: keys map is required")
	}
	for _, key := range this.schema.KeyColumns() {
		if _, ok := keys[key]; !ok {
			return errors.Errorf("validateKeyMap: missing required key '%s'", key)
		}
	}
	return nil
}

func (this *BaseRepositoryImpl) validateGetOneColumnsAndFilter(
	columns []string, filter dmodel.DynamicFields,
) *ft.ClientErrorItem {
	for _, col := range columns {
		if _, ok := this.schema.Column(col); !ok {
			return ft.NewValidationError(
				col,
				ft.ErrorKey("err_unknown_schema_field"),
				"field is not defined on this schema",
			)
		}
	}
	if filter == nil {
		return nil
	}
	keys := make([]string, 0, len(filter))
	for k := range filter {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, ok := this.schema.Column(k); !ok {
			return ft.NewValidationError(
				k,
				ft.ErrorKey("err_unknown_schema_field"),
				"field is not defined on this schema",
			)
		}
	}
	return nil
}

func (this *BaseRepositoryImpl) ensureTenantKeyIn(values dmodel.DynamicFields) error {
	key := this.schema.TenantKey()
	if key == "" {
		return nil
	}
	if _, ok := values[key]; !ok {
		return errors.Errorf("ensureTenantKeyIn: missing tenant key '%s'", key)
	}
	return nil
}

func (this *BaseRepositoryImpl) mergeFilterIntoGraph(
	graph *dmodel.SearchGraph, filter []dmodel.DynamicFields,
) (*dmodel.SearchGraph, error) {
	if len(filter) == 0 {
		if key := this.schema.TenantKey(); key != "" {
			return nil, errors.Errorf(
				"mergeFilterIntoGraph: filter required for tenant-scoped schema, must contain '%s'", key)
		}
		return graph, nil
	}
	f := filter[0]
	if err := this.ensureTenantKeyIn(f); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	filterNodes := make([]dmodel.SearchNode, 0, len(keys))
	for _, k := range keys {
		n := (&dmodel.SearchNode{}).Condition(util.ToPtr(dmodel.NewCondition(k, dmodel.Equals, f[k])))
		filterNodes = append(filterNodes, *n)
	}
	merged := graph
	merged.And(filterNodes...)
	return merged, nil
}

func isNilAnyValue(val any) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

func (this *BaseRepositoryImpl) shouldIncludeEqualFilterField(field string, val any) bool {
	if isNilAnyValue(val) {
		return false
	}
	_, ok := this.schema.Column(field)
	return ok
}

func missingKeyColumnNames(found map[string]bool, keyColumns []string) []string {
	missing := make([]string, 0, len(keyColumns))
	for _, k := range keyColumns {
		if !found[k] {
			missing = append(missing, k)
		}
	}
	return missing
}

// buildEqualNodes builds an Equals SearchNode for each defined, non-nil filter field.
// Unknown keys and nil values are ignored. Key columns must still be present (non-nil).
func (this *BaseRepositoryImpl) buildEqualNodes(filter dmodel.DynamicFields) ([]dmodel.SearchNode, error) {
	keyColumns := this.schema.KeyColumns()
	found := make(map[string]bool, len(keyColumns))
	for _, k := range keyColumns {
		found[k] = false
	}

	nodes := make([]dmodel.SearchNode, 0, len(filter))
	for field, val := range filter {
		if !this.shouldIncludeEqualFilterField(field, val) {
			continue
		}
		n := (&dmodel.SearchNode{}).Condition(util.ToPtr(dmodel.NewCondition(field, dmodel.Equals, val)))
		nodes = append(nodes, *n)
		if _, isKey := found[field]; isKey {
			found[field] = true
		}
	}

	missing := missingKeyColumnNames(found, keyColumns)
	if len(missing) > 0 {
		return nil, errors.Errorf(
			"buildEqualNodes: missing required key columns: %s", strings.Join(missing, ", "))
	}
	return nodes, nil
}

func (this *BaseRepositoryImpl) buildFindOneGraph(
	filter dmodel.DynamicFields, includeArchived bool,
) (*dmodel.SearchGraph, error) {
	allNodes, err := this.buildEqualNodes(filter)
	if err != nil {
		return &dmodel.SearchGraph{}, err
	}
	if !includeArchived {
		archiveCond := dmodel.NewCondition(basemodel.FieldArchivedAt, dmodel.IsNotSet)
		allNodes = append(allNodes, *(&dmodel.SearchNode{}).Condition(util.ToPtr(archiveCond)))
	}
	g := &dmodel.SearchGraph{}
	g.And(allNodes...)
	return g, nil
}

func (this *BaseRepositoryImpl) queryAndScan(ctx corectx.Context, query string) ([]dmodel.DynamicFields, error) {
	rows, err := this.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []dmodel.DynamicFields
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(dmodel.DynamicFields)
		for i, col := range columns {
			val := convertDbValue(values[i])
			if val != nil {
				row[col] = val
			}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func convertDbValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return v
	}
}
