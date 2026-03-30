package baserepo

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
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

func NewBaseRepositoryImpl(param NewBaseRepositoryParam) dyn.BaseRepository {
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
var _ dyn.BaseRepository = (*BaseRepositoryImpl)(nil)

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

// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
func (this *BaseRepositoryImpl) CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[[][]string], error,
) {
	uniqueKeysToCheck := this.filterUniqueKeysWithValues(data)
	if len(uniqueKeysToCheck) == 0 {
		return &dyn.OpResult[[][]string]{Data: nil, HasData: false}, nil
	}

	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlCheckUniqueCollisions(this.schema, uniqueKeysToCheck, data)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[[][]string]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(sqlQuery.Sql)
	rows, err := this.client.Query(ctx, sqlQuery.Sql, sqlQuery.Args...)
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
	return &dyn.OpResult[[][]string]{Data: collidingKeys, HasData: len(collidingKeys) != 0}, nil
}

func (this *BaseRepositoryImpl) DeleteOne(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	if err := this.validateKeyMap(keys); err != nil {
		return nil, err
	}
	// if err := this.ensureTenantKeyIn(keys); err != nil {
	// 	return nil, err
	// }
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlDeleteEqual(this.schema, keys)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	result, err := this.client.Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseRepositoryImpl) Exists(
	ctx corectx.Context, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	if len(keys) == 0 {
		return &dyn.OpResult[dyn.RepoExistsResult]{
			Data:    dyn.RepoExistsResult{},
			HasData: true,
		}, nil
	}
	sqlRes, qbClientErrs, err := this.queryBuilder.SqlExistsMany(this.schema, keys)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.RepoExistsResult]{ClientErrors: *qbClientErrs}, nil
	}
	if sqlRes == nil {
		return &dyn.OpResult[dyn.RepoExistsResult]{
			Data:    dyn.RepoExistsResult{},
			HasData: true,
		}, nil
	}
	return this.runExistsManyQuery(ctx, *sqlRes, keys)
}

func (this *BaseRepositoryImpl) runExistsManyQuery(
	ctx corectx.Context, sqlData orm.SqlExistsManyData, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	this.logQuery(sqlData.Sql)
	rows, err := this.client.Query(ctx, sqlData.Sql, sqlData.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanExistsManyRows(rows, keys)
}

func scanExistsManyRows(rows *sql.Rows, keys []dmodel.DynamicFields) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	var existing, notExisting []dmodel.DynamicFields
	i := 0
	for rows.Next() {
		var flag int
		if err := rows.Scan(&flag); err != nil {
			return nil, err
		}
		if i >= len(keys) {
			break
		}
		if flag == 1 {
			existing = append(existing, keys[i])
		} else {
			notExisting = append(notExisting, keys[i])
		}
		i++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := dyn.RepoExistsResult{Existing: existing, NotExisting: notExisting}
	return &dyn.OpResult[dyn.RepoExistsResult]{Data: out, HasData: true}, nil
}

// Insert inserts a record to database. If the schema defines "created_at", sets current UTC timestamp.
// On success, Data holds the inserted field map; HasData is true when Data is non-nil.
func (this *BaseRepositoryImpl) Insert(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[int], error,
) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	// this.trySetCreatedAt(data)
	// this.trySetEtag(data)
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlInsert(this.schema, data)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	result, err := this.client.Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

// Implements BaseRepository interface
func (this *BaseRepositoryImpl) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[dmodel.DynamicFields], error,
) {
	if vErr := this.validateGetOneColumnsAndFilter(param.Columns, param.Filter); vErr != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: ft.ClientErrors{*vErr}}, nil
	}
	// if err := this.ensureTenantKeyIn(param.Filter); err != nil {
	// 	return nil, err
	// }
	graph, err := this.buildFindOneGraph(param.Filter)
	if err != nil {
		return nil, err
	}
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(
		this.schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
			Columns: param.Columns,
		})
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	rows, err := this.queryAndScan(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{HasData: false}, nil
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: rows[0], HasData: true}, nil
}

func (this *BaseRepositoryImpl) ManageM2m(ctx corectx.Context, destSchemaName string,
	associations []dyn.RepoM2mAssociation, desociations []dyn.RepoM2mAssociation,
) (*dyn.OpResult[int], error) {
	if len(associations) == 0 && len(desociations) == 0 {
		return &dyn.OpResult[int]{Data: 0, HasData: false}, nil
	}
	link, ok := this.schema.M2mPeerLinkForDest(destSchemaName)
	if !ok {
		return nil, errors.Errorf(
			"ManageM2m: no M2M relation from '%s' to '%s'", this.schema.Name(), destSchemaName,
		)
	}
	total := 0
	if len(associations) > 0 {
		insertRes, err := this.insertJunctionRows(ctx, link, associations)
		if err != nil {
			return nil, err
		}
		if len(insertRes.ClientErrors) > 0 {
			return &dyn.OpResult[int]{ClientErrors: insertRes.ClientErrors}, nil
		}
		total += insertRes.Data
	}
	if len(desociations) > 0 {
		deleteRes, err := this.deleteJunctionRows(ctx, link, desociations)
		if err != nil {
			return nil, err
		}
		if len(deleteRes.ClientErrors) > 0 {
			return &dyn.OpResult[int]{ClientErrors: deleteRes.ClientErrors}, nil
		}
		total += deleteRes.Data
	}
	return &dyn.OpResult[int]{Data: total, HasData: true}, nil
}

func (this *BaseRepositoryImpl) deleteJunctionRows(ctx corectx.Context,
	link *dmodel.M2mPeerLink, desociations []dyn.RepoM2mAssociation,
) (*dyn.OpResult[int], error) {
	rows, clientErrs := this.assocToJuncRowArr(link, desociations)
	if len(clientErrs) > 0 {
		return &dyn.OpResult[int]{ClientErrors: clientErrs}, nil
	}
	sqlStr, qbClientErrs, err := this.queryBuilder.SqlDeleteOrAndEquals(link.ThroughSchema, rows)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}
	this.logQuery(*sqlStr)
	execRes, err := this.client.Exec(ctx, *sqlStr)
	if err != nil {
		return nil, err
	}
	n, err := execRes.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseRepositoryImpl) insertJunctionRows(ctx corectx.Context,
	link *dmodel.M2mPeerLink, associations []dyn.RepoM2mAssociation,
) (*dyn.OpResult[int], error) {
	rows, clientErrs := this.assocToJuncRowArr(link, associations)
	if len(clientErrs) > 0 {
		return &dyn.OpResult[int]{ClientErrors: clientErrs}, nil
	}
	sqlRes, qbClientErrs, err := this.queryBuilder.SqlInsertBulk(link.ThroughSchema, rows)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}
	this.logQuery(*sqlRes)
	execRes, err := this.client.Exec(ctx, *sqlRes)
	if err != nil {
		return nil, err
	}
	n, err := execRes.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseRepositoryImpl) assocToJuncRowArr(
	link *dmodel.M2mPeerLink, associations []dyn.RepoM2mAssociation,
) ([]dmodel.DynamicFields, ft.ClientErrors) {
	var errs ft.ClientErrors
	out := make([]dmodel.DynamicFields, 0, len(associations))
	for i := range associations {
		row, rowErrs := this.assocToJuncRow(link, i, associations[i])
		if rowErrs.Count() > 0 {
			for _, e := range rowErrs {
				errs.Append(e)
			}
			continue
		}
		out = append(out, row)
	}
	if errs.Count() > 0 {
		return nil, errs
	}
	return out, nil
}

func (this *BaseRepositoryImpl) assocToJuncRow(
	link *dmodel.M2mPeerLink, idx int, assoc dyn.RepoM2mAssociation,
) (dmodel.DynamicFields, ft.ClientErrors) {
	var errs ft.ClientErrors
	prefix := fmt.Sprintf("associations[%d]", idx)
	if assoc.SrcKeys == nil || assoc.DestKeys == nil {
		errs.Append(*ft.NewValidationError(
			prefix,
			"common.err_missing_required_field",
			"src and dest key maps are required",
		))
		return nil, errs
	}
	appendMissingKeyErrors(&errs, prefix+".srcKeys", assoc.SrcKeys, this.schema.KeyColumns())
	appendMissingKeyErrors(&errs, prefix+".destKeys", assoc.DestKeys, link.DestSchema.KeyColumns())
	if errs.Count() > 0 {
		return nil, errs
	}
	return this.materializeM2mJunctionRow(link, assoc), nil
}

func (this *BaseRepositoryImpl) materializeM2mJunctionRow(
	link *dmodel.M2mPeerLink, assoc dyn.RepoM2mAssociation,
) dmodel.DynamicFields {
	row := make(dmodel.DynamicFields)
	for _, k := range this.schema.PrimaryKeys() {
		row[dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, k)] = assoc.SrcKeys[k]
	}
	if tk := this.schema.TenantKey(); tk != "" {
		col := dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, tk)
		row[col] = assoc.SrcKeys[tk]
	}
	for _, k := range link.DestSchema.PrimaryKeys() {
		row[dmodel.PrefixedThroughColumn(link.DestFieldPrefix, k)] = assoc.DestKeys[k]
	}
	return row
}

func appendMissingKeyErrors(errs *ft.ClientErrors, fieldPrefix string, keys dmodel.DynamicFields, required []string) {
	for _, k := range required {
		if _, ok := keys[k]; !ok || keys[k] == nil {
			errs.Append(*dmodel.NewMissingFieldErr(fieldPrefix + "." + k))
		}
	}
}

// Search fetches records matching the SearchGraph criteria.
// When the schema is tenant-scoped, filter must be provided and contain the tenant key.
// Data uses PagedResult: Total is from COUNT when Size > 0, otherwise len(Items).
func (this *BaseRepositoryImpl) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error,
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
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: countClientErrs,
		}, nil
	}
	rows, scanClientErrs, err := this.runSelectGraphScan(ctx, merged, param)
	if err != nil {
		return nil, err
	}
	if len(scanClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: scanClientErrs,
		}, nil
	}
	if size <= 0 {
		total = len(rows)
	}
	paged := dyn.PagedResultData[dmodel.DynamicFields]{
		Items: rows,
		Total: total,
		Page:  page,
		Size:  size,
	}
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    paged,
		HasData: len(rows) != 0,
	}, nil
}

func (this *BaseRepositoryImpl) countRowsMatchingGraph(
	ctx corectx.Context, graph *dmodel.SearchGraph,
) (int, ft.ClientErrors, error) {
	qbRes, qbClientErrs, err := this.queryBuilder.SqlCountGraph(this.schema, dmodel.GetSchemaRegistry(), graph)
	if err != nil {
		return 0, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return 0, *qbClientErrs, nil
	}
	countSql := *qbRes
	this.logQuery(countSql)
	n, ierr := this.queryScalarInt(ctx, countSql)
	return n, nil, ierr
}

func (this *BaseRepositoryImpl) runSelectGraphScan(
	ctx corectx.Context, graph *dmodel.SearchGraph, param dyn.RepoSearchParam,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(
		this.schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
			Columns: param.Columns,
			Page:    param.Page,
			Size:    param.Size,
		})
	if err != nil {
		return nil, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, *qbClientErrs, nil
	}

	this.logQuery(*sqlQuery)
	rows, err := this.queryAndScan(ctx, *sqlQuery)
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

// Update updates a record. The data map must contain primary keys and tenant key.
// If the schema defines "updated_at", sets current UTC timestamp.
func (this *BaseRepositoryImpl) Update(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[dmodel.DynamicFields], error,
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
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlUpdateEqual(this.schema, data, filters)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	_, err = this.client.Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: data, HasData: true}, nil
}

func (this *BaseRepositoryImpl) execUpdate(
	ctx corectx.Context, data dmodel.DynamicFields, filters dmodel.DynamicFields,
) (sql.Result, ft.ClientErrors, error) {
	qbRes, qbClientErrs, err := this.queryBuilder.SqlUpdateEqual(this.schema, data, filters)
	if err != nil {
		return nil, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, *qbClientErrs, nil
	}
	query := *qbRes

	this.logQuery(query)
	sqlResult, err := this.client.Exec(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	return sqlResult, nil, nil
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

// func (this *BaseRepositoryImpl) trySetArchivedAt(data dmodel.DynamicFields) error {
// 	if _, ok := this.schema.Column(basemodel.FieldArchivedAt); !ok {
// 		return errors.Errorf(
// 			"Archive: schema '%s' does not define field '%s'", this.schema.Name(), basemodel.FieldArchivedAt)
// 	}
// 	field := this.schema.MustField(basemodel.FieldArchivedAt)
// 	data[basemodel.FieldArchivedAt] = *field.DataType().DefaultValue().Get()
// 	return nil
// }

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
			delete(data, key)
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
		n := dmodel.NewSearchNode().NewCondition(k, dmodel.Equals, f[k])
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
		n := dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, val)
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

func (this *BaseRepositoryImpl) buildFindOneGraph(filter dmodel.DynamicFields) (*dmodel.SearchGraph, error) {
	allNodes, err := this.buildEqualNodes(filter)
	if err != nil {
		return &dmodel.SearchGraph{}, err
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
