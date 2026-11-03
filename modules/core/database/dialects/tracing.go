package dialects

import (
	"database/sql"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// tracedSpanOptions keep SQL out of spans (BE-04): no statement text, and otelsql records bind
// arguments only through an AttributesGetter, which is never set. The span name is the driver
// operation (sql.conn.query, sql.conn.exec, ...), which carries no query either.
var tracedSpanOptions = otelsql.SpanOptions{
	DisableQuery:         true,
	DisableErrSkip:       true,
	OmitConnResetSession: true,
	OmitRows:             true,
}

// openTraced is sql.Open through otelsql. Pool metrics (otelsql.RegisterDBStatsMetrics) are not
// registered: this rollout exports traces only.
func openTraced(driverName string, dsn string) (*sql.DB, error) {
	return otelsql.Open(driverName, dsn,
		otelsql.WithSpanOptions(tracedSpanOptions),
		otelsql.WithAttributes(dbSystemAttr(driverName)),
	)
}

func dbSystemAttr(driverName string) attribute.KeyValue {
	switch driverName {
	case orm.DialectPostgres:
		return semconv.DBSystemPostgreSQL
	case orm.DialectMySql:
		return semconv.DBSystemMySQL
	default:
		return semconv.DBSystemKey.String(driverName)
	}
}
