package database

import (
	"context"
	"database/sql"
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicentity/model"
	orm "github.com/sky-as-code/nikki-erp/common/dynamicentity/orm"
	eschema "github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"go.bryk.io/pkg/errors"
)

type DynamicRepositoryBase[TEntity any] struct {
	dbConn       *DbConnection
	dbEntity     *orm.DbEntity
	queryBuilder *orm.PgQueryBuilder
}

// NewDynamicRepositoryBase creates a new DynamicRepositoryBase instance for the given schema name.
func NewDynamicRepositoryBase[TEntity any](
	dbConn *DbConnection, schemaName string, moduleName string, subModName ...string,
) (*DynamicRepositoryBase[TEntity], error) {
	schema, ok := eschema.GetSchema(schemaName, moduleName, subModName...)
	if !ok {
		return nil, fmt.Errorf("schema '%s' not found", schemaName)
	}

	tableName := schema.TableName()
	if tableName == "" {
		return nil, fmt.Errorf("schema '%s' has no table name set", schemaName)
	}

	dbEntity, err := orm.NewPgDbEntity(schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to create db entity for schema '%s': %w", schemaName, err)
	}

	return &DynamicRepositoryBase[TEntity]{
		dbConn:       dbConn,
		dbEntity:     dbEntity,
		queryBuilder: orm.NewPgQueryBuilder(dbEntity),
	}, nil
}

// CreateM creates a new record from an EntityMap.
// Returns an EntityMap containing only Primary & Tenant Keys.
func (this *DynamicRepositoryBase[TEntity]) CreateM(ctx context.Context, data dmodel.EntityMap) (dmodel.EntityMap, error) {
	sqlQuery, err := this.queryBuilder.SqlInsertMap(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate insert SQL")
	}

	result, err := this.dbConn.Db.ExecContext(ctx, sqlQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute insert query")
	}

	// Extract primary and tenant keys from the input data
	output := make(dmodel.EntityMap)
	for _, pk := range this.dbEntity.PrimaryKeys {
		if value, ok := data[pk]; ok {
			output[pk] = value
		}
	}
	if this.dbEntity.TenantKey != nil {
		if value, ok := data[*this.dbEntity.TenantKey]; ok {
			output[*this.dbEntity.TenantKey] = value
		}
	}

	// If no primary keys were provided, try to get last insert ID
	if len(output) == 0 {
		lastID, err := result.LastInsertId()
		if err == nil && lastID > 0 {
			// Assume first primary key is auto-increment
			if len(this.dbEntity.PrimaryKeys) > 0 {
				output[this.dbEntity.PrimaryKeys[0]] = lastID
			}
		}
	}

	return output, nil
}

// CreateS creates a new record from a struct.
// Uses StructToEntityMap to convert the struct to EntityMap, then invokes CreateM.
// Returns the created entity as TEntity.
func (this *DynamicRepositoryBase[TEntity]) CreateS(ctx context.Context, data TEntity) (*TEntity, error) {
	entityMap := dmodel.StructToEntityMap(data)
	resultMap, err := this.CreateM(ctx, entityMap)
	if err != nil {
		return nil, err
	}
	result := dmodel.EntityMapToStruct[TEntity](resultMap)
	return result, nil
}

// CreateBulkM creates multiple records from EntityMaps.
func (this *DynamicRepositoryBase[TEntity]) CreateBulkM(ctx context.Context, data []dmodel.EntityMap) ([]dmodel.EntityMap, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	sqlQuery, err := this.queryBuilder.SqlInsertBulkMaps(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate bulk insert SQL")
	}

	_, err = this.dbConn.Db.ExecContext(ctx, sqlQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute bulk insert query")
	}

	// Extract primary and tenant keys from each input data
	output := make([]dmodel.EntityMap, len(data))
	for i, row := range data {
		outputRow := make(dmodel.EntityMap)
		for _, pk := range this.dbEntity.PrimaryKeys {
			if value, ok := row[pk]; ok {
				outputRow[pk] = value
			}
		}
		if this.dbEntity.TenantKey != nil {
			if value, ok := row[*this.dbEntity.TenantKey]; ok {
				outputRow[*this.dbEntity.TenantKey] = value
			}
		}
		output[i] = outputRow
	}

	return output, nil
}

// CreateBulkS creates multiple records from structs.
// Uses StructToEntityMap to convert each struct to EntityMap, then invokes CreateBulkM.
// Returns the created entities as []TEntity.
func (this *DynamicRepositoryBase[TEntity]) CreateBulkS(ctx context.Context, data []TEntity) ([]TEntity, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	entityMaps := make([]dmodel.EntityMap, len(data))
	for i, item := range data {
		entityMaps[i] = dmodel.StructToEntityMap(item)
	}

	resultMaps, err := this.CreateBulkM(ctx, entityMaps)
	if err != nil {
		return nil, err
	}

	results := make([]TEntity, len(resultMaps))
	for i, resultMap := range resultMaps {
		entity := dmodel.EntityMapToStruct[TEntity](resultMap)
		results[i] = *entity
	}

	return results, nil
}

// UpdateM updates a record by primary key from an EntityMap.
func (this *DynamicRepositoryBase[TEntity]) UpdateM(ctx context.Context, data dmodel.EntityMap) error {
	sqlQuery, err := this.queryBuilder.SqlUpdateByPkMap(data)
	if err != nil {
		return errors.Wrap(err, "failed to generate update SQL")
	}

	result, err := this.dbConn.Db.ExecContext(ctx, sqlQuery)
	if err != nil {
		return errors.Wrap(err, "failed to execute update query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated")
	}

	return nil
}

// UpdateS updates a record by primary key from a struct.
// Uses StructToEntityMap to convert the struct to EntityMap, then invokes UpdateM.
func (this *DynamicRepositoryBase[TEntity]) UpdateS(ctx context.Context, data TEntity) error {
	entityMap := dmodel.StructToEntityMap(data)
	return this.UpdateM(ctx, entityMap)
}

// Delete deletes records matching the filters.
// Returns the number of rows deleted.
func (this *DynamicRepositoryBase[TEntity]) Delete(ctx context.Context, filters map[string]string) (int64, error) {
	// Convert map[string]string to EntityMap
	filterMap := make(dmodel.EntityMap)
	for k, v := range filters {
		filterMap[k] = v
	}

	sqlQuery, err := this.queryBuilder.SqlDeleteEqualStruct(filterMap)
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate delete SQL")
	}

	result, err := this.dbConn.Db.ExecContext(ctx, sqlQuery)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute delete query")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get rows affected")
	}

	return rowsAffected, nil
}

// Exists checks if a record exists matching the filters.
func (this *DynamicRepositoryBase[TEntity]) Exists(ctx context.Context, filters map[string]string) (bool, error) {
	searchGraph := this.buildSearchGraphFromFilters(filters)
	columns := this.getDefaultColumns()

	sqlQuery, err := this.queryBuilder.SqlSelectGraph(searchGraph, columns)
	if err != nil {
		return false, errors.Wrap(err, "failed to generate select SQL")
	}

	// Use LIMIT 1 to optimize the query
	limitQuery := sqlQuery + " LIMIT 1"
	rows, err := this.dbConn.Db.QueryContext(ctx, limitQuery)
	if err != nil {
		return false, errors.Wrap(err, "failed to execute select query")
	}
	defer rows.Close()

	return rows.Next(), nil
}

// SearchS searches for records matching the filters and returns them as []TEntity.
func (this *DynamicRepositoryBase[TEntity]) SearchS(ctx context.Context, filters map[string]string) ([]TEntity, error) {
	entityMaps, err := this.SearchM(ctx, filters)
	if err != nil {
		return nil, err
	}

	results := make([]TEntity, len(entityMaps))
	for i, entityMap := range entityMaps {
		entity := dmodel.EntityMapToStruct[TEntity](entityMap)
		results[i] = *entity
	}

	return results, nil
}

// SearchM searches for records matching the filters and returns them as []EntityMap.
func (this *DynamicRepositoryBase[TEntity]) SearchM(ctx context.Context, filters map[string]string) ([]dmodel.EntityMap, error) {
	searchGraph := this.buildSearchGraphFromFilters(filters)
	columns := this.getAllColumns()

	sqlQuery, err := this.queryBuilder.SqlSelectGraph(searchGraph, columns)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate select SQL")
	}

	rows, err := this.dbConn.Db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute select query")
	}
	defer rows.Close()

	return this.scanRowsToEntityMaps(rows, columns)
}

// buildSearchGraphFromFilters builds a SearchGraph from a map of filters.
func (this *DynamicRepositoryBase[TEntity]) buildSearchGraphFromFilters(filters map[string]string) orm.SearchGraph {
	filterMap := make(dmodel.EntityMap)
	for k, v := range filters {
		filterMap[k] = v
	}

	conditions := make([]orm.Condition, 0, len(filterMap))
	for field, value := range filterMap {
		conditions = append(conditions, orm.NewCondition(field, orm.Equals, value))
	}

	searchGraph := orm.SearchGraph{
		And: make([]orm.SearchNode, len(conditions)),
	}
	for i, cond := range conditions {
		searchGraph.And[i] = orm.SearchNode{
			Condition: &cond,
		}
	}

	return searchGraph
}

// getDefaultColumns returns the default columns to select (primary keys and tenant key).
// Used for Exists checks where we only need to verify existence.
func (this *DynamicRepositoryBase[TEntity]) getDefaultColumns() []string {
	columns := append([]string{}, this.dbEntity.PrimaryKeys...)
	if this.dbEntity.TenantKey != nil {
		columns = append(columns, *this.dbEntity.TenantKey)
	}
	return columns
}

// getAllColumns returns all column names from the entity.
// Used for SearchM and SearchS to return complete records.
func (this *DynamicRepositoryBase[TEntity]) getAllColumns() []string {
	columns := make([]string, len(this.dbEntity.Columns))
	for i, col := range this.dbEntity.Columns {
		columns[i] = col.Name
	}
	return columns
}

// scanRowsToEntityMaps scans database rows into EntityMap slices.
func (this *DynamicRepositoryBase[TEntity]) scanRowsToEntityMaps(rows *sql.Rows, columns []string) ([]dmodel.EntityMap, error) {
	var results []dmodel.EntityMap

	for rows.Next() {
		// Create a slice to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		// Convert scanned values to EntityMap
		entityMap := make(dmodel.EntityMap)
		for i, col := range columns {
			entityMap[col] = values[i]
		}

		results = append(results, entityMap)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return results, nil
}
