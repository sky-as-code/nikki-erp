package baserepo

import (
	"sort"
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const (
	columnCreatedAt = "created_at"
	columnUpdatedAt = "updated_at"
	columnArchiveAt = "archive_at"
	columnEtag      = "etag"
)

type NewBaseRepositoryParam struct {
	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	Logger       logging.LoggerService
	QueryBuilder orm.QueryBuilder
	Schema       *schema.EntitySchema
}

func NewBaseRepositoryImpl(param NewBaseRepositoryParam) dEnt.BaseRepository {
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
var _ dEnt.BaseRepository = (*BaseRepositoryImpl)(nil)

type BaseRepositoryImpl struct {
	client          orm.DbClient
	queryBuilder    orm.QueryBuilder
	logger          logging.LoggerService
	schema          *schema.EntitySchema
	sqlDebugEnabled bool
}

func (this *BaseRepositoryImpl) GetSchema() *schema.EntitySchema {
	return this.schema
}

// Insert inserts a record. If the entity defines "created_at", sets current UTC timestamp.
// Returns a map of primary keys and tenant key of the created record, or nil on error.
func (this *BaseRepositoryImpl) Insert(ctx dEnt.Context, data schema.DynamicFields) (schema.DynamicFields, error) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	this.trySetCreatedAt(data)
	this.trySetEtag(data)
	query, err := this.queryBuilder.SqlInsert(this.schema, data)
	if err != nil {
		return nil, err
	}

	this.logQuery(query)
	_, err = this.client.Exec(ctx, query)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Update updates a record. The data map must contain primary keys and tenant key.
// If the entity defines "updated_at", sets current UTC timestamp.
func (this *BaseRepositoryImpl) Update(ctx dEnt.Context, data schema.DynamicFields) (schema.DynamicFields, error) {
	if err := this.ensureTenantKeyIn(data); err != nil {
		return nil, err
	}
	this.trySetUpdatedAt(data)
	this.trySetEtag(data)
	query, err := this.queryBuilder.SqlUpdateByPk(this.schema, data)
	if err != nil {
		return nil, err
	}

	this.logQuery(query)
	_, err = this.client.Exec(ctx, query)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindByPk fetches a single record by primary keys and tenant key.
// Returns (nil, nil) when record is not found and no error occurred.
func (this *BaseRepositoryImpl) FindByPk(ctx dEnt.Context, keys schema.DynamicFields) (schema.DynamicFields, error) {
	if err := this.validateKeyMap(keys); err != nil {
		return nil, err
	}
	if err := this.ensureTenantKeyIn(keys); err != nil {
		return nil, err
	}
	graph := this.buildPkSearchGraph(keys)
	query, err := this.queryBuilder.SqlSelectGraph(this.schema, graph, nil)
	if err != nil {
		return nil, err
	}

	this.logQuery(query)
	rows, err := this.queryAndScan(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// Search fetches records matching the SearchGraph criteria.
// When the entity is tenant-scoped, filter must be provided and contain the tenant key.
// Returns nil when no records found.
func (this *BaseRepositoryImpl) Search(
	ctx dEnt.Context, graph schema.SearchGraph, columns []string, filter ...schema.DynamicFields,
) ([]schema.DynamicFields, error) {
	merged, err := this.mergeFilterIntoGraph(graph, filter)
	if err != nil {
		return nil, err
	}
	query, err := this.queryBuilder.SqlSelectGraph(this.schema, merged, columns)
	if err != nil {
		return nil, err
	}

	this.logQuery(query)
	return this.queryAndScan(ctx, query)
}

// Archive sets archive_at to current UTC timestamp for the record identified by keys.
// Returns error if the entity does not define "archive_at" column.
func (this *BaseRepositoryImpl) Archive(ctx dEnt.Context, keys schema.DynamicFields) (schema.DynamicFields, error) {
	if _, ok := this.schema.Column(columnArchiveAt); !ok {
		return nil, errors.Errorf("entity '%s' does not define column '%s'", this.schema.Name(), columnArchiveAt)
	}
	record, err := this.FindByPk(ctx, keys)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	record[columnArchiveAt] = time.Now().UTC()
	updated, err := this.Update(ctx, record)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// Delete removes the record identified by primary keys and tenant key.
// Returns the number of deleted records.
func (this *BaseRepositoryImpl) Delete(ctx dEnt.Context, keys schema.DynamicFields) (int64, error) {
	if err := this.validateKeyMap(keys); err != nil {
		return 0, err
	}
	if err := this.ensureTenantKeyIn(keys); err != nil {
		return 0, err
	}
	query, err := this.queryBuilder.SqlDeleteEqual(this.schema, keys)
	if err != nil {
		return 0, err
	}

	this.logQuery(query)
	result, err := this.client.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
func (this *BaseRepositoryImpl) CheckUniqueCollisions(ctx dEnt.Context, data schema.DynamicFields) ([][]string, error) {
	uniqueKeysToCheck := this.filterUniqueKeysWithValues(data)
	if len(uniqueKeysToCheck) == 0 {
		return nil, nil
	}

	query, args, err := this.queryBuilder.SqlCheckUniqueCollisions(this.schema, uniqueKeysToCheck, data)
	if err != nil {
		return nil, err
	}

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
	return collidingKeys, rows.Err()
}

func (this *BaseRepositoryImpl) logQuery(query string) {
	if this.sqlDebugEnabled {
		this.logger.Debugf(query)
	}
}

func (this *BaseRepositoryImpl) filterUniqueKeysWithValues(data schema.DynamicFields) [][]string {
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

func (this *BaseRepositoryImpl) trySetCreatedAt(data schema.DynamicFields) {
	if _, ok := this.schema.Column(columnCreatedAt); !ok {
		return
	}
	if data != nil {
		data[columnCreatedAt] = time.Now().UTC()
	}
}

func (this *BaseRepositoryImpl) trySetUpdatedAt(data schema.DynamicFields) {
	if _, ok := this.schema.Column(columnUpdatedAt); !ok {
		return
	}
	if data != nil {
		data[columnUpdatedAt] = time.Now().UTC()
	}
}

func (this *BaseRepositoryImpl) trySetEtag(data schema.DynamicFields) {
	if _, ok := this.schema.Column(columnEtag); !ok {
		return
	}
	if data != nil {
		data[columnEtag] = *model.NewEtag()
	}
}

func (this *BaseRepositoryImpl) extractKeyMap(data schema.DynamicFields) schema.DynamicFields {
	if data == nil {
		return nil
	}
	result := make(schema.DynamicFields)
	for _, key := range this.schema.KeyColumns() {
		if v, ok := data[key]; ok {
			result[key] = v
		}
	}
	return result
}

func (this *BaseRepositoryImpl) validateKeyMap(keys schema.DynamicFields) error {
	if len(keys) == 0 {
		return errors.New("keys map is required")
	}
	for _, key := range this.schema.KeyColumns() {
		if _, ok := keys[key]; !ok {
			return errors.Errorf("missing required key '%s'", key)
		}
	}
	return nil
}

func (this *BaseRepositoryImpl) ensureTenantKeyIn(values schema.DynamicFields) error {
	key := this.schema.TenantKey()
	if key == "" {
		return nil
	}
	if _, ok := values[key]; !ok {
		return errors.Errorf("missing tenant key '%s'", key)
	}
	return nil
}

func (this *BaseRepositoryImpl) mergeFilterIntoGraph(
	graph schema.SearchGraph, filter []schema.DynamicFields,
) (schema.SearchGraph, error) {
	if len(filter) == 0 {
		if key := this.schema.TenantKey(); key != "" {
			return schema.SearchGraph{}, errors.Errorf("filter required for tenant-scoped entity, must contain '%s'", key)
		}
		return graph, nil
	}
	f := filter[0]
	if err := this.ensureTenantKeyIn(f); err != nil {
		return schema.SearchGraph{}, err
	}
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	filterNodes := make([]schema.SearchNode, 0, len(keys))
	for _, k := range keys {
		n := (&schema.SearchNode{}).Condition(util.ToPtr(schema.NewCondition(k, schema.Equals, f[k])))
		filterNodes = append(filterNodes, *n)
	}
	merged := graph
	(&merged).And(filterNodes...)
	return merged, nil
}

func (this *BaseRepositoryImpl) buildPkSearchGraph(keys schema.DynamicFields) schema.SearchGraph {
	nodes := make([]schema.SearchNode, 0, len(this.schema.KeyColumns()))
	for _, key := range this.schema.KeyColumns() {
		if v, ok := keys[key]; ok {
			n := (&schema.SearchNode{}).Condition(util.ToPtr(schema.NewCondition(key, schema.Equals, v)))
			nodes = append(nodes, *n)
		}
	}
	g := &schema.SearchGraph{}
	g.And(nodes...)
	return *g
}

func (this *BaseRepositoryImpl) queryAndScan(ctx dEnt.Context, query string) ([]schema.DynamicFields, error) {
	rows, err := this.client.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []schema.DynamicFields
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(schema.DynamicFields)
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
