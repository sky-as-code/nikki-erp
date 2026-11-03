package dialects

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const fakeDriverName = "dialects-fake"

// fakeDriver accepts any statement and returns an empty result, so the test needs no database.
type fakeDriver struct{}

type fakeConn struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return fakeConn{}, nil }

func (fakeConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (fakeConn) Close() error                        { return nil }
func (fakeConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (fakeConn) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func init() {
	sql.Register(fakeDriverName, fakeDriver{})
}

// BE-04: a DB span names the operation only; neither the statement nor a bind value is recorded.
func TestDbSpansCarryNoQueryTextOrArgs(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	conn, err := openTraced(fakeDriverName, "")
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.ExecContext(context.Background(), "UPDATE iam_users SET password_hash = $1", "s3cret-value")
	require.NoError(t, err)

	spans := recorder.Ended()
	require.NotEmpty(t, spans)
	for _, span := range spans {
		assert.NotContains(t, span.Name(), "iam_users")
		for _, attr := range span.Attributes() {
			value := attr.Value.Emit()
			assert.False(t, strings.Contains(value, "iam_users") || strings.Contains(value, "s3cret"),
				"span %s attribute %s leaks %q", span.Name(), attr.Key, value)
		}
	}
}
