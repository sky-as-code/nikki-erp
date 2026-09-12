package repository

import (
	"github.com/huandu/go-sqlbuilder"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// tenantIdField is the column the multi-tenant build adds to every table. The nikkierp binary has
// no tenant key at all, so it is absent there and the scoping below becomes a no-op.
const tenantIdField = "tenant_id"

// nextStreamSeqSql draws the next value of the stream sequence.
//
// The sequence name is a literal rather than a parameter: nextval() takes a regclass, and a bound
// parameter would arrive as text and be rejected. It is a compile-time constant of this package,
// never caller input, so there is nothing here to inject.
const nextStreamSeqSql = `SELECT nextval('notification_stream_seq')`

// nowUtcSqlExpr stamps a read in UTC.
//
// The database clock rather than the application's: two instances with drifting clocks would
// otherwise write read timestamps that disagree about which of two reads came first.
const nowUtcSqlExpr = `(NOW() AT TIME ZONE 'UTC')`

// scopeSelectToTenant adds the tenant predicate to a hand-built read.
//
// The multi-tenant repository decorator adds this automatically to everything that goes through the
// engine, but these statements do not, so the scope has to be explicit. Without it an inbox read
// would cross tenant boundaries.
func scopeSelectToTenant(ctx corectx.Context, builder *sqlbuilder.SelectBuilder, alias string) {
	tenantId := actingTenantId(ctx)
	if tenantId == "" {
		return
	}
	builder.Where(builder.Equal(alias+"."+tenantIdField, tenantId))
}

func scopeUpdateToTenant(ctx corectx.Context, builder *sqlbuilder.UpdateBuilder) {
	tenantId := actingTenantId(ctx)
	if tenantId == "" {
		return
	}
	builder.Where(builder.Equal(tenantIdField, tenantId))
}

// actingTenantId reads the tenant the request is acting in, or empty in a build that has no
// tenants. The domain constraints are where the engine puts it.
func actingTenantId(ctx corectx.Context) string {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil {
		return ""
	}
	raw, found := constraints[tenantIdField]
	if !found || raw == nil {
		return ""
	}
	return tenantIdString(raw)
}

// tenantIdString accepts the several shapes an id takes as it crosses the dynamic-model boundary.
func tenantIdString(raw any) string {
	switch typed := raw.(type) {
	case string:
		return typed
	case *string:
		if typed == nil {
			return ""
		}
		return *typed
	case interface{ String() string }:
		return typed.String()
	}
	return ""
}
