package orm

import (
	"context"
	"database/sql"
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
