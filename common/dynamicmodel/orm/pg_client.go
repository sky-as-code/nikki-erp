package orm

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// PgClient implements DBClient for PostgreSQL.
type PgClient struct {
	db *sql.DB
}

// Ensure interface implementation at compile time.
var _ DbClient = (*PgClient)(nil)

func NewPgClient(db *sql.DB) DbClient {
	return &PgClient{db: db}
}

// Exec executes a SQL statement and returns the result.
func (this *PgClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return this.db.ExecContext(ctx, query, args...)
}

// Query executes a SQL query and returns the rows.
func (this *PgClient) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return this.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SQL query that returns at most one row.
func (this *PgClient) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return this.db.QueryRowContext(ctx, query, args...)
}

// BeginTx starts a transaction.
func (this *PgClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (DbTxClient, error) {
	tx, err := this.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &pgTxClient{db: this.db, tx: tx}, nil
}

// Close releases the database connection pool.
func (this *PgClient) Close() error {
	return this.db.Close()
}

// Ensure pgTxClient implements DBTxClient at compile time.
var _ DbTxClient = (*pgTxClient)(nil)

// pgTxClient implements DBTxClient for PostgreSQL transactions.
type pgTxClient struct {
	db *sql.DB
	tx *sql.Tx

	callbacksMu sync.Mutex
	callbacks   []func()
}

// ErrTxNested is returned when BeginTx is called on a client already in a transaction.
var ErrTxNested = errors.New("pgClient: cannot start a transaction while in a transaction")

func (this *pgTxClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return this.tx.ExecContext(ctx, query, args...)
}

func (this *pgTxClient) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return this.tx.QueryContext(ctx, query, args...)
}

func (this *pgTxClient) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return this.tx.QueryRowContext(ctx, query, args...)
}

func (this *pgTxClient) BeginTx(_ context.Context, _ *sql.TxOptions) (DbTxClient, error) {
	return nil, ErrTxNested
}

func (this *pgTxClient) Close() error {
	return this.db.Close()
}

func (this *pgTxClient) Commit() error {
	err := this.tx.Commit()
	callbacks := this.takeCallbacks()
	if err != nil {
		return err
	}
	for _, fn := range callbacks {
		runCommittedCallback(fn)
	}
	return nil
}

func (this *pgTxClient) Rollback() error {
	this.takeCallbacks()
	return this.tx.Rollback()
}

func (this *pgTxClient) OnCommitted(fn func()) {
	if fn == nil {
		return
	}
	this.callbacksMu.Lock()
	defer this.callbacksMu.Unlock()
	this.callbacks = append(this.callbacks, fn)
}

func (this *pgTxClient) takeCallbacks() []func() {
	this.callbacksMu.Lock()
	defer this.callbacksMu.Unlock()
	callbacks := this.callbacks
	this.callbacks = nil
	return callbacks
}

// runCommittedCallback isolates a callback: the data is already committed, so a panicking side
// effect must neither fail the caller's request nor skip the callbacks queued after it.
func runCommittedCallback(fn func()) {
	defer func() {
		if err := ft.RecoverPanic(recover(), "OnCommitted callback panicked"); err != nil {
			slog.Error(err.Error())
		}
	}()
	fn()
}
