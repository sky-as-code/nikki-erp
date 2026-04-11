package orm

import (
	"context"
	"database/sql"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type DbClient interface {
	// Exec executes a SQL statement (INSERT, UPDATE, DELETE) and returns the result.
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)

	// Query executes a SQL query (SELECT) and returns the rows.
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// QueryRow executes a SQL query that returns at most one row.
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row

	// BeginTx starts a transaction. Caller must call Commit or Rollback on the returned client.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (DbTxClient, error)

	// Close releases any resources held by the client.
	Close() error
}

// DbTxClient extends DBClient for use within a transaction.
// It adds Commit and Rollback. BeginTx returns an error when called on a client
// that is already in a transaction.
type DbTxClient interface {
	DbClient
	Commit() error
	Rollback() error
}

// SelectColumn is a select-list token for graph queries: plain field path, DISTINCT::field, or aggregates.
type SelectColumn string

// SqlSelectGraphOpts holds optional parameters for SqlSelectGraph.
type SqlSelectGraphOpts struct {
	// Columns limits which columns to fetch; empty means all columns (*).
	Columns []SelectColumn
	// Page is the 0-based page index used to compute the OFFSET (OFFSET = Page * Size).
	// Ignored when Size is 0.
	Page int
	// Size sets the LIMIT clause. 0 means no limit/offset is applied.
	Size int
}

// SqlCheckUniqueCollisionsData holds parameterized SQL and arguments from SqlCheckUniqueCollisions.
type SqlCheckUniqueCollisionsData struct {
	Sql  string
	Args []any
}

// SqlExistsManyData holds parameterized SQL and arguments from SqlExistsMany.
type SqlExistsManyData struct {
	Sql  string
	Args []any
}

type QueryBuilder interface {
	// SqlCreateTable returns DDL strings: [0] = CREATE TABLE, [1..] = CREATE UNIQUE INDEX per PartialUnique.
	SqlCreateTable(schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry) ([]string, *ft.ClientErrors, error)
	SqlSelectGraph(schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts) (
		*string, *ft.ClientErrors, error)
	// SqlExistsGraph builds SELECT EXISTS (SELECT 1 ...) with the same FROM/JOIN/WHERE as SqlCountGraph.
	SqlExistsGraph(schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph) (
		*string, *ft.ClientErrors, error)
	// SqlCountGraph builds SELECT COUNT(*) with the same WHERE as SqlSelectGraph (no ORDER BY, LIMIT, OFFSET).
	SqlCountGraph(schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph) (
		*string, *ft.ClientErrors, error)
	// SqlInsert builds INSERT for one row. When onConflictPkDoNothing is true, appends
	// ON CONFLICT (<schema primary keys>) DO NOTHING.
	SqlInsert(schema *dmodel.ModelSchema, data dmodel.DynamicFields, ignoreConflict bool) (
		*string, *ft.ClientErrors, error)
	// SqlInsertBulk builds a multi-row INSERT. When onConflictPkDoNothing is true, appends
	// ON CONFLICT (<schema primary keys>) DO NOTHING.
	SqlInsertBulk(schema *dmodel.ModelSchema, rows []dmodel.DynamicFields, ignoreConflict bool) (
		*string, *ft.ClientErrors, error)
	SqlUpdateEqual(schema *dmodel.ModelSchema, data dmodel.DynamicFields, filters dmodel.DynamicFields) (
		*string, *ft.ClientErrors, error)
	// SqlDeleteEqual generates a DELETE statement with the given filters using only equal predicates.
	// This DELETE statement can result in one or multiple rows being deleted.
	SqlDeleteEqual(schema *dmodel.ModelSchema, filters dmodel.DynamicFields) (*string, *ft.ClientErrors, error)
	// SqlDeleteOrAndEquals deletes rows matching (AND of equals per map) OR (next map AND...) ...
	SqlDeleteOrAndEquals(schema *dmodel.ModelSchema, filters []dmodel.DynamicFields) (*string, *ft.ClientErrors, error)
	// SqlCheckUniqueCollisions builds SQL that returns 1 per row where the unique key has a collision, else 0.
	// Input: uniqueKeysToCheck - subset of dmodel.AllUniques() where data has all values (no nil).
	// Data.Sql / Data.Args are for execution. Result rows are single int: 1 = collision, 0 = no collision.
	// Returns (nil, nil, nil) when uniqueKeysToCheck is empty.
	SqlCheckUniqueCollisions(
		schema *dmodel.ModelSchema, uniqueKeysToCheck [][]string, data dmodel.DynamicFields,
	) (*SqlCheckUniqueCollisionsData, *ft.ClientErrors, error)
	// SqlExistsMany builds SQL that returns one row per key set: 1 if a row exists matching all columns, else 0.
	// Each key map must use the same column set; column order follows sorted keys from the first map.
	// Returns (nil, nil, nil) when keys is empty.
	SqlExistsMany(schema *dmodel.ModelSchema, keys []dmodel.DynamicFields) (*SqlExistsManyData, *ft.ClientErrors, error)
	// RegisterPredefinedPredicate binds a filter field name to a treatment. Omit schemaName to apply to all schemas
	// (PredefinedPredicateAllSchemas). Operator uses dmodel.Operator values from model/database condition operators.
	RegisterPredefinedPredicate(fieldName string, treatment PredefinedPredicateTreatment, schemaName ...string) error
	GetPredefinedPredicate(fieldName string, schemaName string) PredefinedPredicateTreatment
}
