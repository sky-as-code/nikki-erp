package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"
)

// A minimal database/sql driver: just enough to open a real *sql.Tx whose Commit can be made to
// fail, so the callback queue is tested against the standard library rather than a stub.

var committedTestDriverSeq atomic.Int64

type committedTestDriver struct{ commitErr error }

func (this *committedTestDriver) Open(string) (driver.Conn, error) {
	return &committedTestConn{commitErr: this.commitErr}, nil
}

type committedTestConn struct{ commitErr error }

func (this *committedTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("not supported")
}
func (this *committedTestConn) Close() error { return nil }
func (this *committedTestConn) Begin() (driver.Tx, error) {
	return &committedTestTx{err: this.commitErr}, nil
}

type committedTestTx struct{ err error }

func (this *committedTestTx) Commit() error   { return this.err }
func (this *committedTestTx) Rollback() error { return nil }

func beginCommittedTestTx(t *testing.T, commitErr error) DbTxClient {
	name := fmt.Sprintf("orm-committed-test-%d", committedTestDriverSeq.Add(1))
	sql.Register(name, &committedTestDriver{commitErr: commitErr})
	db, err := sql.Open(name, "")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	tx, err := NewPgClient(db).BeginTx(context.Background(), nil)
	require.NoError(t, err)
	return tx
}

func TestOnCommittedRunsCallbacksInOrderAfterCommit(t *testing.T) {
	tx := beginCommittedTestTx(t, nil)
	var calls []int
	tx.OnCommitted(func() { calls = append(calls, 1) })
	tx.OnCommitted(func() { calls = append(calls, 2) })

	assert.Empty(t, calls, "nothing runs before the commit")
	require.NoError(t, tx.Commit())
	assert.Equal(t, []int{1, 2}, calls)
}

func TestOnCommittedPanicIsIsolated(t *testing.T) {
	tx := beginCommittedTestTx(t, nil)
	ranAfter := false
	tx.OnCommitted(func() { panic("boom") })
	tx.OnCommitted(func() { ranAfter = true })

	require.NoError(t, tx.Commit(), "a panicking side effect does not fail a committed write")
	assert.True(t, ranAfter, "the callbacks queued after a panic still run")
}

func TestOnCommittedSkippedOnRollback(t *testing.T) {
	tx := beginCommittedTestTx(t, nil)
	ran := false
	tx.OnCommitted(func() { ran = true })

	require.NoError(t, tx.Rollback())
	assert.False(t, ran)
}

func TestOnCommittedSkippedWhenCommitFails(t *testing.T) {
	tx := beginCommittedTestTx(t, errors.New("serialization failure"))
	ran := false
	tx.OnCommitted(func() { ran = true })

	require.Error(t, tx.Commit())
	assert.False(t, ran)
}
