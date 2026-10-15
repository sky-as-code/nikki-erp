package baserepo

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// A filter on an SQL-computed field that reads ${ctx.*} renders its subquery in the COUNT and in
// the page SELECT alike. Search must hand the request context to both, on every path, or the
// builder refuses the query with err_computed_context_missing.

// recordingBuilder captures the options Search gives the count and the page select. The count
// succeeds (answered by zeroCountClient); the select then stops the search with a client error,
// so no page ever needs scanning.
type recordingBuilder struct {
	orm.QueryBuilder
	countOpts  []orm.SqlSelectGraphOpts
	selectOpts []orm.SqlSelectGraphOpts
}

func (this *recordingBuilder) SqlCountGraph(
	_ *dmodel.ModelSchema, _ *dmodel.SchemaRegistry, _ *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	this.countOpts = append(this.countOpts, opts)
	query := "SELECT 0"
	return &query, nil, nil
}

func (this *recordingBuilder) SqlSelectGraph(
	_ *dmodel.ModelSchema, _ *dmodel.SchemaRegistry, _ *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts,
) (*string, *ft.ClientErrors, error) {
	this.selectOpts = append(this.selectOpts, opts)
	return nil, &ft.ClientErrors{{Key: "stop", Message: "recorded"}}, nil
}

// projectingRecorder also claims nested projection, steering Search onto the projected-edge path.
type projectingRecorder struct {
	recordingBuilder
}

func (this *projectingRecorder) PlanNestedProjection(
	_ *dmodel.ModelSchema, _ *dmodel.SchemaRegistry, _ []orm.SelectColumn,
) (*orm.NestedProjection, *ft.ClientErrors, error) {
	return &orm.NestedProjection{}, nil, nil
}

func TestComputedContext_PlainSearchForwardsToCountAndSelect(t *testing.T) {
	builder := &recordingBuilder{}
	searchWith(t, builder, []string{"id", "sku"})
	assertContextForwarded(t, builder)
}

func TestComputedContext_NestedColumnSearchForwardsToCountAndSelect(t *testing.T) {
	builder := &recordingBuilder{}
	searchWith(t, builder, []string{"id", "peer.name"})
	assertContextForwarded(t, builder)
}

func TestComputedContext_ProjectedEdgeSearchForwardsToCountAndSelect(t *testing.T) {
	builder := &projectingRecorder{}
	searchWith(t, builder, []string{"id", "peer.name"})
	assertContextForwarded(t, &builder.recordingBuilder)
}

var requestContext = map[string]any{"location_id": "LOC1"}

func searchWith(t *testing.T, builder orm.QueryBuilder, fields []string) {
	t.Helper()
	repo := virtualRepo(t)
	repo.queryBuilder = builder
	repo.client = zeroCountClient(t)
	result, err := repo.Search(nil, dyn.RepoSearchParam{
		Fields:          fields,
		Page:            0,
		Size:            10,
		ComputedContext: requestContext,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotEmpty(t, result.ClientErrors, "the recording select must stop the search")
	assert.Equal(t, "stop", result.ClientErrors[0].Key)
}

func assertContextForwarded(t *testing.T, builder *recordingBuilder) {
	t.Helper()
	require.Len(t, builder.countOpts, 1)
	require.Len(t, builder.selectOpts, 1)
	assert.Equal(t, requestContext, builder.countOpts[0].ComputedContext, "count")
	assert.Equal(t, requestContext, builder.selectOpts[0].ComputedContext, "select")
}

// ---- a database that answers every query with a single zero ----------------------------------

type zeroDbClient struct {
	db *sql.DB
}

func (this zeroDbClient) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, errors.New("zeroDbClient: Exec is not supported")
}

func (this zeroDbClient) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return this.db.QueryContext(context.Background(), query, args...)
}

func (this zeroDbClient) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return this.db.QueryRowContext(context.Background(), query, args...)
}

func (this zeroDbClient) BeginTx(ctx context.Context, opts *sql.TxOptions) (orm.DbTxClient, error) {
	return nil, errors.New("zeroDbClient: transactions are not supported")
}

func (this zeroDbClient) Close() error {
	return this.db.Close()
}

var registerZeroDriver sync.Once

func zeroCountClient(t *testing.T) orm.DbClient {
	t.Helper()
	registerZeroDriver.Do(func() { sql.Register("baserepo_zero", zeroDriver{}) })
	db, err := sql.Open("baserepo_zero", "")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return zeroDbClient{db: db}
}

type zeroDriver struct{}

func (zeroDriver) Open(string) (driver.Conn, error) { return zeroConn{}, nil }

type zeroConn struct{}

func (zeroConn) Prepare(string) (driver.Stmt, error) { return zeroStmt{}, nil }
func (zeroConn) Close() error                        { return nil }
func (zeroConn) Begin() (driver.Tx, error)           { return nil, errors.New("not supported") }

type zeroStmt struct{}

func (zeroStmt) Close() error  { return nil }
func (zeroStmt) NumInput() int { return -1 }
func (zeroStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("not supported")
}
func (zeroStmt) Query([]driver.Value) (driver.Rows, error) { return &zeroRows{}, nil }

type zeroRows struct {
	done bool
}

func (this *zeroRows) Columns() []string { return []string{"count"} }
func (this *zeroRows) Close() error      { return nil }
func (this *zeroRows) Next(dest []driver.Value) error {
	if this.done {
		return io.EOF
	}
	dest[0] = int64(0)
	this.done = true
	return nil
}
